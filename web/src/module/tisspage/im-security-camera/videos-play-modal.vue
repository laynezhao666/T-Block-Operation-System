
<template>
  <el-dialog
    :visible.sync="visible"
    title="摄像头视频"
    width="800px"
    :close-on-click-modal="false"
    append-to-body
  >
    <div v-if="!cameraList || cameraList.length === 0" class="empty-tip">
      {{ emptyAlertText }}
    </div>
    <div v-else class="video-list">
      <div
        v-for="(camera, index) in cameraList"
        :key="index"
        class="video-item"
      >
        <div class="camera-name">{{ camera.name || camera.cameraName || `摄像头${index + 1}` }}</div>
        <div class="video-placeholder">
          <i class="el-icon-video-camera" />
          <span>视频流暂不可用</span>
        </div>
      </div>
    </div>
  </el-dialog>
</template>

<script>
export default {
  name: 'VideosPlayModal',
  props: {
    forceReplaceInfo: {
      type: Boolean,
      default: false,
    },
    emptyAlertText: {
      type: String,
      default: '暂无关联摄像头',
    },
  },
  data() {
    return {
      visible: false,
      cameraList: [],
    };
  },
  methods: {
    show(cameraList) {
      this.cameraList = cameraList || [];
      this.visible = true;
    },
    hide() {
      this.visible = false;
    },
  },
};
</script>

<style scoped>
.empty-tip {
  text-align: center;
  padding: 40px 0;
  color: #909399;
  font-size: 14px;
}

.video-list {
  display: flex;
  flex-wrap: wrap;
  gap: 16px;
}

.video-item {
  flex: 1;
  min-width: 300px;
  max-width: 50%;
}

.camera-name {
  font-size: 14px;
  font-weight: 500;
  margin-bottom: 8px;
  color: #303133;
}

.video-placeholder {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  height: 200px;
  background: #f5f7fa;
  border-radius: 4px;
  color: #909399;
}

.video-placeholder i {
  font-size: 48px;
  margin-bottom: 8px;
}

.video-placeholder span {
  font-size: 12px;
}
</style>
