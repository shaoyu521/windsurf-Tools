<script setup lang="ts">
const props = defineProps<{
  eyebrow: string
  value: string
  caption: string
  accent?: 'sky' | 'emerald' | 'amber' | 'rose'
}>()
</script>

<template>
  <article class="metric-card" :class="`metric-card--${props.accent || 'sky'}`">
    <p class="metric-card__eyebrow">{{ props.eyebrow }}</p>
    <strong class="metric-card__value">{{ props.value }}</strong>
    <p class="metric-card__caption">{{ props.caption }}</p>
  </article>
</template>

<style scoped>
.metric-card {
  position: relative;
  overflow: hidden;
  border-radius: 24px;
  border: 1px solid rgba(255, 255, 255, 0.08);
  background:
    linear-gradient(180deg, rgba(255, 255, 255, 0.05), rgba(11, 16, 24, 0.92)),
    radial-gradient(circle at top left, rgba(255, 255, 255, 0.08), transparent 50%);
  padding: 1.1rem 1.15rem;
  min-height: 150px;
  display: grid;
  align-content: space-between;
  gap: 0.7rem;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.06);
}

.metric-card::after {
  content: '';
  position: absolute;
  inset: auto -18% -34% auto;
  width: 140px;
  height: 140px;
  border-radius: 999px;
  opacity: 0.4;
  filter: blur(8px);
}

.metric-card--sky::after {
  background: radial-gradient(circle, rgba(69, 168, 255, 0.72), transparent 70%);
}

.metric-card--emerald::after {
  background: radial-gradient(circle, rgba(47, 210, 166, 0.72), transparent 70%);
}

.metric-card--amber::after {
  background: radial-gradient(circle, rgba(255, 173, 77, 0.72), transparent 70%);
}

.metric-card--rose::after {
  background: radial-gradient(circle, rgba(255, 117, 146, 0.72), transparent 70%);
}

.metric-card__eyebrow {
  margin: 0;
  font-size: 0.72rem;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: var(--wt-text-soft);
}

.metric-card__value {
  font-size: clamp(2rem, 2.6vw, 2.7rem);
  line-height: 0.92;
  letter-spacing: -0.06em;
  color: var(--wt-text-strong);
}

.metric-card__caption {
  margin: 0;
  max-width: 14rem;
  font-size: 0.96rem;
  color: var(--wt-text-muted);
}
</style>
