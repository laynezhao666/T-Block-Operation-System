
<template>
  <el-modal
    :visible.sync="modalVisible"
    title="版本特性"
  >
    <template slot="title">
      版本特性
    </template>
    <el-block class="timeline">
      <el-timeline>
        <el-timeline-item
          v-for="(log, index) in logs"
          :key="index"
          :icon="log.icon"
          :type="log.type"
          :size="log.size"
          :timestamp="log.timestamp"
          placement="top"
        >
          <slot>
            <el-collapse
              v-model="activeName"
              accordion
            >
              <el-collapse-item :name="index">
                <template slot="title">
                  <div class="title">
                    {{ log.title }}
                  </div>
                  <div class="sub-title">
                    <span>{{ log.subTitle }}</span>
                  </div>
                </template>
                <el-card>
                  <ol
                    v-for="(content, index) in log.contents"
                    :key="index"
                    class="content"
                  >
                    {{ index + 1 }}、{{ content.content }}
                    <ol
                      v-if="content.details"
                      style="list-style-type:disc;font-weight:normal;font-size:14px;margin-left:40px"
                    >
                      <li
                        v-for="(detail, index) in content.details"
                        :key="index"
                      >
                        {{ detail.content }}
                        <ol
                          v-if="detail.details"
                          style="list-style-type:circle;margin-left:40px"
                        >
                          <li
                            v-for="(detail, index) in detail.details"
                            :key="index"
                          >
                            {{ detail.content }}
                            <ul
                              v-if="detail.details"
                              style="list-style-type:square;margin-left:40px"
                            >
                              <li
                                v-for="(detail, index) in detail.details"
                                :key="index"
                              >
                                {{ detail.content }}
                              </li>
                            </ul>
                          </li>
                        </ol>
                      </li>
                    </ol>
                  </ol>
                </el-card>
                <!-- <el-button type="text">
                  查看更多
                </el-button> -->
              </el-collapse-item>
            </el-collapse>
          </slot>
        </el-timeline-item>
      </el-timeline>
    </el-block>
  </el-modal>
</template>

<script>
import { changeLogs } from '../const/change-log';

export default {
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      logs: changeLogs,
      activeName: 0,
    };
  },
  computed: {
    modalVisible: {
      set(v) {
        this.$emit('update:visible', v);
      },
      get() {
        return this.visible;
      },
    },
  },
};
</script>

<style lang="scss" scoped>
/deep/ .el-block {
   height: 100%;
   .el-block__body {
    height: 100%;
    .el-block__body-inner {
      height: 100%;
    }
  }
}

.timeline {
  padding: 30px;
  height: calc(100% - 64px);
  overflow: auto;

  // padding-left: 120px;
  /deep/ .el-timeline-item__timestamp.is-top {
    // position: absolute;
    // top: 0;
    // left: -90px;
    // margin: 0;
    // padding: 0;
  }
  /deep/ .el-collapse {
    border: none;
    .el-collapse-item__header {
      border: none;
      padding-top: 0px;
      padding-bottom: 30px;
      height: 100%;
    }
    .el-collapse-item__wrap {
      border-bottom: none;
    }
  }
  .title {
    color: #333;
    font-size: 20px;
    font-weight: 400;
  }
  .sub-title {
    color: #999;
    font-size: 13px;
    padding: 4px;
  }
  .content {
    line-height: 30px;
    font-weight: bold;
    font-size: 15px;
    // background-color: rgba(0,0,0,0.1);
    // padding: 0 10px;
  }
  li::marker {
    // font-size: 20px;
    // line-height: 33.5px;
    // height: 33px;
  }
}
</style>
