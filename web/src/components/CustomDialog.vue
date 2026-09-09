<template>
  <el-dialog :modelValue="visible" style="min-width: 300px; max-width: 500px" :before-close="close" align-center>
    <template #header>
      <h4 class="modal-title">
        <slot name="title"></slot>
      </h4>
    </template>
    <div style="font-size: 18px">
      <slot name="text" ></slot>
    </div>
    <template #footer>
      <span class="dialog-footer">
        <el-button @click="close">{{ cancelText || $t('dialog.cancel') }}</el-button>
        <el-button id="btn_confirm" data-testid="dialog-confirm" type="primary" @click="yes" v-if="confirmEnabled">{{ $t('dialog.confirm') }}</el-button>
      </span>
    </template>
  </el-dialog>
</template>

<script>
export default {
  name: 'CustomDialog',
  emits: ['confirm', 'cancel'],
  props: {
    visible: Boolean,
    confirmEnabled: {
      type: Boolean,
      default: true
    },
    cancelText: {
      type: String,
      default: null
    }
  },
  methods: {
    yes () {
      this.$emit('confirm')
    },
    close () {
      this.$emit('cancel')
    }
  }
}
</script>
