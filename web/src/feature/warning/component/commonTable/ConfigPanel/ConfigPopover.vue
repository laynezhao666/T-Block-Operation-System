<template>
  <div>
    <import-modal
      :visible.sync="showModal"
      :text="text"
      @success="uploadSuccess"
    />
    <el-dropdown-item
      v-if="hasRights('canImport')"
      v-appmatrixauth="roles.write"
      icon="tn-icon-upload"
      @click.native="show"
    >
      导入
    </el-dropdown-item>
    <el-dropdown-item
      v-if="hasRights('canExport')"
      icon="tn-icon tn-icon-import"
      @click.native="download"
    >
      导出全部
    </el-dropdown-item>
  </div>
</template>
<script>
import ImportModal from './ImportModal';
import mixin from '../script/mixin';
import { eventBus } from '../script/eventBus';

export default {
  components: { ImportModal },
  inject: ['configCgi', 'tableConfig'],
  mixins: [mixin],
  props: {
    text: {
      type: String,
      required: true,
    },
    codes: {
      type: Object,
      required: true,
    },
    roles: {
      type: Object,
      required: true,
    },
  },
  data() {
    return {
      showModal: false,
      rights: this.tableConfig.rights,
      analysis: this.tableConfig.analysis || false,
    };
  },
  mounted() {
    eventBus.$on('toggleAnalysisDownload', () => {
      this.downloadAnalysis();
    });
  },
  beforeDestroy() {
    eventBus.$off('toggleAnalysisDownload');
  },
  methods: {
    uploadSuccess() {
      this.$emit('uploadSuccess');
    },
    show() {
      this.showModal = true;
    },
    download() {
      this.$emit('export');
    },
    downloadAnalysis() {
      this.$emit('export', [], 'analysis');
    },
    analysisHistory() {
      this.$emit('doelse', { type: 'analysis' });
    },
  },
};
</script>
