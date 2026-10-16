<template>
  <div class="px-4 py-3 border-b border-border shrink-0">
    <div class="flex items-center space-x-2">
      <div class="w-6 h-6 rounded-full bg-primary-light flex items-center justify-center flex-shrink-0">
        <span class="text-primary font-semibold text-xs">O</span>
      </div>
      <div class="flex-1 min-w-0">
        <div class="flex items-center space-x-1">
          <span class="text-xs font-semibold text-text truncate">
            {{ organization?.name || 'Personal' }}
          </span>
          <button
            class="w-1.5 h-1.5 rounded-full bg-success flex-shrink-0"
            title="Connected"
          ></button>
        </div>
        <p class="text-[10px] text-text-muted truncate">
          {{ organization?.description || 'Default workspace' }}
        </p>
      </div>
      <button
        class="w-6 h-6 rounded-full hover:bg-hover flex items-center justify-center text-text-muted transition-colors"
        title="Workspace settings"
      >
        <svg
          class="w-4 h-4"
          fill="none"
          viewBox="0 0 24 24"
          stroke="currentColor"
        >
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M10.325 4.317c.426-1.756 2.924-1.756 3.35 0a1.724 1.724 0 002.573 1.066c1.543-.94 3.31.826 2.37 2.37a1.724 1.724 0 001.065 2.572c1.756.426 1.756 2.924 0 3.35a1.724 1.724 0 00-1.066 2.573c.94 1.543-.826 3.31-2.37 2.37a1.724 1.724 0 00-2.572 1.065c-.426 1.756-2.924 1.756-3.35 0a1.724 1.724 0 00-2.573-1.066c-1.543.94-3.31-.826-2.37-2.37a1.724 1.724 0 00-1.065-2.572c-1.756-.426-1.756-2.924 0-3.35a1.724 1.724 0 001.066-2.573c-.94-1.543.826-3.31 2.37-2.37.996.608 2.296.07 2.572-1.065z" />
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M15 12a3 3 0 11-6 0 3 3 0 016 0z"
          />
        </svg>
      </button>
    </div>
  </div>
</template>

<script setup lang="ts">
import type { OrganizationView } from '@/types'

interface Props {
  organization: OrganizationView | null
}

const props = defineProps<Props>()
</script>

<style scoped></style>
