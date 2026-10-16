<template>
  <div
    v-if="isOpen"
    class="fixed inset-0 z-50 flex items-center justify-center"
  >
    <div
      class="absolute inset-0 bg-black/50 backdrop-blur-sm transition-opacity"
      @click="closeSettings"
    ></div>
    <div
      class="bg-surface w-full max-w-md rounded-xl border border-border shadow-xl mx-4 p-6 animate-slide-up flex-shrink-0"
    >
      <h3 class="text-lg font-semibold text-text mb-6">Settings</h3>
      
      <div class="space-y-5">
        <div>
          <label class="block text-xs font-semibold text-text-muted uppercase tracking-wider mb-2">
            Display name
          </label>
          <input
            type="text"
            v-model="displayName"
            class="w-full bg-input border border-border rounded-lg px-3 py-2 text-sm text-text focus:border-primary focus:ring-1 focus:ring-primary outline-none"
          />
        </div>

        <div class="pt-2 flex justify-end space-x-2">
          <button
            @click="closeSettings"
            class="px-4 py-2 text-sm text-text-muted hover:text-text transition-colors"
          >
            Cancel
          </button>
          <button
            @click="saveDisplayName"
            class="px-4 py-2 text-sm font-semibold text-primary hover:text-primary-dark transition-colors"
            :disabled="!displayName.trim() || isSaving"
          >
            {{ isSaving ? 'Saving...' : 'Save' }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'

interface Props {
  isOpen: boolean
  currentName: string
  saveFn: (name: string) => Promise<void>
  isSaving: boolean
}

const props = defineProps<Props>()

const emit = defineEmits(['close', 'save'])

const displayName = ref(props.currentName)

const canSave = computed(() => {
  return displayName.value.trim().length > 0 && !props.isSaving
})

function closeSettings() {
  emit('close')
}

async function saveDisplayName() {
  if (!displayName.value.trim()) return
  await props.saveFn(displayName.value.trim())
  emit('save')
}
</script>

<style scoped>
.animate-slide-up {
  animation: slideUp 0.3s ease-out forwards;
}

@keyframes slideUp {
  from {
    opacity: 0;
    transform: translateY(20px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}
</style>
