<template>
  <el-dialog
    title="同步时间组"
    :visible.sync="visible"
    width="30%"
  >
    <div
      v-loading="!syncResultMap"
      element-loading-text="同步中，请等候。"
      class="sync-info"
    >
      <div
        class="main-text"
      >
        <span class="success-text">
          {{ syncResultMap && syncResultMap.success || 0 }}
        </span>
        个控制器操作成功，
        <span class="fail-text">
          {{ syncResultMap && syncResultMap.fail || 0 }}
        </span>
        个控制器操作失败
      </div>
      <div
        v-if="syncResultMap && syncResultMap.fail"
        class="fail-alert"
      >
        操作失败的控制器可能不在线或未启用
      </div>
    </div>

    <span
      slot="footer"
    >
      <el-button
        type="primary"
        @click="close"
      >确定</el-button>
    </span>
  </el-dialog>
</template>

<script>
import getEdgeRequest from 'feature/utils/request';

export default {
  data() {
    return {
      controlIds: null,
      syncResultMap: null,
    };
  },
  computed: {
    visible: {
      get() {
        return Boolean(this.controlIds);
      },
      set(v) {
        if (!v) {
          this.close();
        }
      },
    },
  },
  methods: {
    startSync(controlIds = []) {
      this.controlIds = controlIds;
      this.requestRemote();
    },
    close() {
      this.controlIds = null;
      this.syncResultMap = null;
    },
    async requestRemote() {
      const { controlIds } = this;

      const resp = await getEdgeRequest(this.$axios).post('/api/dcos/tdac-cgi/time-groups/sync', {
        sync_all: controlIds?.length ? 0 : 1,
        controllers: controlIds,
      });

      this.syncResultMap = {
        success: resp.success_count,
        fail: resp.fail_count,
      };
    },
  },
};
</script>

<style lang="scss" scoped>
.main-text {
  font-size: 14px;
  font-weight: 500;
}

.success-text {
  color: var(--tn-color-success);
  font-weight: 600;
}

.fail-text {
  color: var(--tn-color-danger);
  font-weight: 600;
}

.fail-alert {
  color: #999;
  font-size: 12px;
}

.sync-info {
  padding-bottom: 24px;
}
</style>
