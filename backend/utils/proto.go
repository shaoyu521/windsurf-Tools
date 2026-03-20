package utils

// EncodeStringField builds a length-delimited string field
func EncodeStringField(fieldNum uint8, value string) []byte {
	var result []byte
	bytes := []byte(value)
	length := len(bytes)

	// Field tag: (fieldNum << 3) | 2 (wire type 2 = length-delimited)
	result = append(result, (fieldNum<<3)|2)

	// Varint length
	if length < 128 {
		result = append(result, byte(length))
	} else {
		result = append(result, byte((length&0x7F)|0x80))
		result = append(result, byte(length>>7))
	}

	result = append(result, bytes...)
	return result
}

// ReadVarintSimple reads a varint from the current position and returns the value and new position
func ReadVarintSimple(data []byte, pos int) (uint64, int, bool) {
	var result uint64 = 0
	var shift uint = 0
	for pos < len(data) {
		b := data[pos]
		pos++
		result |= uint64(b&0x7F) << shift
		shift += 7
		if (b & 0x80) == 0 {
			return result, pos, true
		}
		if shift >= 64 {
			return 0, pos, false
		}
	}
	return 0, pos, false
}

// FindJWTInProtobuf traverses a simple protobuf message searching for a string that looks like a JWT
func FindJWTInProtobuf(data []byte) (string, bool) {
	pos := 0
	for pos < len(data) {
		tag, newPos, ok := ReadVarintSimple(data, pos)
		if !ok {
			break
		}
		pos = newPos
		wireType := tag & 7

		switch wireType {
		case 0:
			// varint - skip
			_, newPos, ok := ReadVarintSimple(data, pos)
			if !ok {
				return "", false
			}
			pos = newPos
		case 2:
			// length-delimited
			lengthVar, newPos, ok := ReadVarintSimple(data, pos)
			if !ok {
				return "", false
			}
			pos = newPos
			length := int(lengthVar)
			if pos+length > len(data) {
				return "", false
			}
			
			fieldData := data[pos : pos+length]
			s := string(fieldData)
			if len(s) > 100 && len(s) > 3 && s[:3] == "eyJ" {
				return s, true
			}
			
			// Recursive search if embedded message
			if length > 2 {
				if jwt, found := FindJWTInProtobuf(fieldData); found {
					return jwt, true
				}
			}
			
			pos += length
		case 5:
			pos += 4 // fixed32
		case 1:
			pos += 8 // fixed64
		default:
			return "", false
		}
	}
	return "", false
}
