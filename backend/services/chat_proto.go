package services

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"strings"
	"windsurf-tools-wails/backend/utils"
)

// ══════════════════════════════════════════════════════════════
// GetChatMessage protobuf 编解码
// ══════════════════════════════════════════════════════════════

// ChatMessage OpenAI 格式的消息
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// parseProtoFields 从 protobuf 字节流中解析所有顶层字段（复用 windsurf.go 的 protoField）
func parseProtoFields(data []byte) protoMessage {
	return decodeProtoMessage(data)
}

// BuildChatRequest 构造 GetChatMessage 的 protobuf 请求体（不含 gRPC 5 字节信封）。
func BuildChatRequest(messages []ChatMessage, apiKey, jwt, conversationID string) []byte {
	// F1: metadata
	metadata := buildChatMetadata(apiKey, jwt)
	metaField := encodeBytesField(1, metadata)

	// F2: conversation_id
	var convField []byte
	if conversationID != "" {
		convField = utils.EncodeStringField(2, conversationID)
	}

	// F3: chat content — 将 messages 拼接为单个用户提示
	prompt := flattenMessages(messages)
	contentInner := utils.EncodeStringField(1, prompt)
	contentField := encodeBytesField(3, contentInner)

	var body []byte
	body = append(body, metaField...)
	if len(convField) > 0 {
		body = append(body, convField...)
	}
	body = append(body, contentField...)
	return body
}

// WrapGRPCEnvelope 给 protobuf 消息加上 gRPC 5 字节信封
func WrapGRPCEnvelope(payload []byte) []byte {
	envelope := make([]byte, 5+len(payload))
	envelope[0] = 0x00
	binary.BigEndian.PutUint32(envelope[1:5], uint32(len(payload)))
	copy(envelope[5:], payload)
	return envelope
}

const (
	streamEnvelopeCompressed = 0x01
	streamEnvelopeEndStream  = 0x02
)

type streamEnvelope struct {
	Flags   byte
	Payload []byte
}

// ExtractGRPCEnvelopes 从响应字节流中提取原始 envelope（保留 flag，用于处理压缩和 end-stream）。
func ExtractGRPCEnvelopes(data []byte) []streamEnvelope {
	var envelopes []streamEnvelope
	pos := 0
	for pos+5 <= len(data) {
		flags := data[pos]
		payloadLen := int(binary.BigEndian.Uint32(data[pos+1 : pos+5]))
		pos += 5
		if pos+payloadLen > len(data) {
			break
		}
		payload := append([]byte(nil), data[pos:pos+payloadLen]...)
		envelopes = append(envelopes, streamEnvelope{
			Flags:   flags,
			Payload: payload,
		})
		pos += payloadLen
	}
	return envelopes
}

func decodeStreamEnvelopePayload(flags byte, payload []byte) ([]byte, error) {
	if flags&streamEnvelopeCompressed == 0 {
		return append([]byte(nil), payload...), nil
	}
	reader, err := gzip.NewReader(bytes.NewReader(payload))
	if err != nil {
		return nil, fmt.Errorf("gzip reader: %w", err)
	}
	defer reader.Close()
	decoded, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("gzip read: %w", err)
	}
	return decoded, nil
}

// ParseChatResponseChunk 从流式响应的一个 gRPC 帧中提取文本 delta。
// data 应为去掉 5 字节信封后的 protobuf payload。
func ParseChatResponseChunk(data []byte) (text string, isDone bool, err error) {
	fields := parseProtoFields(data)
	if len(fields) == 0 {
		return "", false, fmt.Errorf("empty protobuf chunk")
	}

	// 真实 GetChatMessage 响应的文本增量位于 F6.F3；先走这个路径，
	// 避免把 bot id / request id 之类的 metadata 当成回答正文。
	if preferred := extractProtoTextAtPath(fields, 6, 3); preferred != "" {
		text = preferred
	} else {
		text = extractLegacyChatText(fields)
	}

	for _, f := range fields {
		if f.Number == 2 && f.Wire == 0 && f.Varint != 0 {
			isDone = true
		}
	}
	return text, isDone, nil
}

// ExtractGRPCFrames 从流式响应字节流中提取多个 gRPC 帧。
func ExtractGRPCFrames(data []byte) [][]byte {
	var frames [][]byte
	for _, envelope := range ExtractGRPCEnvelopes(data) {
		if envelope.Flags&streamEnvelopeEndStream != 0 {
			continue
		}
		frame, err := decodeStreamEnvelopePayload(envelope.Flags, envelope.Payload)
		if err != nil {
			continue
		}
		frames = append(frames, frame)
	}
	return frames
}

func extractProtoTextAtPath(message protoMessage, path ...uint64) string {
	if len(path) == 0 {
		return ""
	}
	var sb strings.Builder
	for _, field := range message {
		if field.Number != path[0] || field.Wire != 2 {
			continue
		}
		if len(path) == 1 {
			if isLikelyUTF8(field.Bytes) {
				sb.WriteString(string(field.Bytes))
			}
			continue
		}
		sub := parseProtoFields(field.Bytes)
		if len(sub) == 0 {
			continue
		}
		sb.WriteString(extractProtoTextAtPath(sub, path[1:]...))
	}
	return sb.String()
}

func extractLegacyChatText(fields protoMessage) string {
	var sb strings.Builder
	for _, f := range fields {
		switch {
		case f.Number == 1 && f.Wire == 2:
			if isLikelyUTF8(f.Bytes) {
				s := string(f.Bytes)
				if !looksLikeChatMetadataString(s) {
					sb.WriteString(s)
				}
			} else {
				sub := parseProtoFields(f.Bytes)
				for _, sf := range sub {
					if sf.Wire == 2 && isLikelyUTF8(sf.Bytes) {
						s := string(sf.Bytes)
						if !looksLikeChatMetadataString(s) {
							sb.WriteString(s)
						}
					}
				}
			}
		case f.Number == 3 && f.Wire == 2:
			if isLikelyUTF8(f.Bytes) {
				sb.WriteString(string(f.Bytes))
			}
		}
	}
	return sb.String()
}

func looksLikeChatMetadataString(s string) bool {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "" {
		return false
	}
	if strings.HasPrefix(s, "bot-") || strings.HasPrefix(s, "msg_") || strings.HasPrefix(s, "req_") {
		return true
	}
	if len(s) == 36 && strings.Count(s, "-") == 4 {
		return true
	}
	return false
}

// ── 内部辅助 ──

func buildChatMetadata(apiKey, jwt string) []byte {
	var meta []byte
	meta = append(meta, utils.EncodeStringField(1, WindsurfAppName)...)
	meta = append(meta, utils.EncodeStringField(2, WindsurfVersion)...)
	meta = append(meta, utils.EncodeStringField(3, apiKey)...)
	meta = append(meta, utils.EncodeStringField(4, "en")...)
	meta = append(meta, utils.EncodeStringField(5, "windows")...)
	meta = append(meta, utils.EncodeStringField(7, WindsurfClient)...)
	meta = append(meta, utils.EncodeStringField(12, WindsurfAppName)...)
	if jwt != "" {
		meta = append(meta, utils.EncodeStringField(21, jwt)...)
	}
	return meta
}

func flattenMessages(messages []ChatMessage) string {
	if len(messages) == 0 {
		return ""
	}
	if len(messages) == 1 {
		return messages[0].Content
	}
	var sb strings.Builder
	for _, m := range messages {
		switch m.Role {
		case "system":
			sb.WriteString("[System]\n")
		case "assistant":
			sb.WriteString("[Assistant]\n")
		default:
			sb.WriteString("[User]\n")
		}
		sb.WriteString(m.Content)
		sb.WriteString("\n\n")
	}
	return strings.TrimSpace(sb.String())
}

func encodeBytesField(fieldNum uint64, data []byte) []byte {
	tag := writeVarint((fieldNum << 3) | 2)
	length := writeVarint(uint64(len(data)))
	result := make([]byte, 0, len(tag)+len(length)+len(data))
	result = append(result, tag...)
	result = append(result, length...)
	result = append(result, data...)
	return result
}
