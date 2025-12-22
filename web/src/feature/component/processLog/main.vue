<template>
  <div class="el-process-log">
    <el-block no-padding>
      <template slot="header">
        {{ title }}
      </template>
      <el-collapse
        v-if="logs.length"
        class="i-collapse"
      >
        <el-collapse-item
          v-for="(item, index) in logs"
          :key="index"
          class="i-log"
        >
          <template slot="title">
            <div class="log-item">
              <span class="log-item__index">{{ reverseIndex(index) }}</span>
              <h4 class="log-item__type">
                {{ item.opType }}
              </h4>
              <span
                v-if="item.opUser"
                class="log-item__user"
              >处理人：{{ item.opUser }}</span>
              <span
                v-else
                class="log-item__user"
              >处理人：{{ item.opUsers }}</span>
              <span class="log-item__time">{{ item.opTime }}</span>
            </div>
          </template>
          <el-form label-position="right">
            <slot
              name="item"
              :item="item"
              :index="index"
            >
              <div v-if="item.logType === 'call'">
                <el-form-item
                  :label="item.columns.username||'姓名不详'"
                  :label-width="'160px'"
                >
                  <span style="margin-right: 150px;">{{ item.columns.phone }}</span>
                  <!-- <img :src="srcUrl(item.columns.call_status)"> -->
                  <span
                    id="phoneImgText"
                    class="w80 ml5 mr20"
                  >{{ item.columns.call_status }}</span>
                </el-form-item>
              </div>
              <div v-else-if="item.logType === 'mail_out'||item.logType === 'mail_in'">
                <el-form-item>
                  <div class="mail-info">
                    <div class="mail-header">
                      <span>
                        发件人：
                        <span class="send-from">{{ item.columns.mail_from }}</span>
                      </span>
                      <span class="ml30">
                        收件人：
                        <span class="send-to">{{ item.columns.mail_to }}</span>
                      </span>
                    </div>
                    <div class="mail-content">
                      <table class="w">
                        <caption class="tl fb mail-title">
                          {{ item.columns.mail_subject }}
                        </caption>
                        <tr>
                          <td class="mail-text">
                            <div v-html="item.columns.mail_content" />
                          </td>
                        </tr>
                      </table>
                    </div>
                  </div>
                </el-form-item>
              </div>
              <div v-else-if="typeof(item.columns)==='object'">
                <div
                  v-for="(v, k) in item.columns"
                  :key="k"
                >
                  <el-form-item
                    v-if="v!==''&&v!==null&&v.length"
                    :label="mappingKey(k)"
                    :label-width="'160px'"
                  >
                    <span v-if="mappingKey(k) === '备件信息'">
                      <span
                        v-for="(part, pk) in v"
                        :key="pk"
                      >
                        <span>
                          <span>{{ part.name }}</span>
                          <span>({{ part.model }}、{{ part.sn }})</span>&nbsp;&nbsp;
                        </span>
                      </span>
                    </span>
                    <span v-else-if="mappingKey(k) === '维修工程师' && Array.isArray(v)">
                      <span
                        v-for="(engineer, pk) in v"
                        :key="pk"
                      >
                        <span>
                          <span>{{ engineer.name }}</span>
                          <span>({{ engineer.phone }}、{{ engineer.id }})</span>&nbsp;&nbsp;
                        </span>
                      </span>
                    </span>
                    <span
                      v-else-if="mappingKey(k).indexOf('照片') > -1 || ['事件报告'].indexOf(mappingKey(k)) > -1
                        || mappingKey(k).indexOf('附件') > -1"
                      style="word-break: break-all;"
                    >
                      <span
                        v-for="(rep, pk) in JSON.parse(v)"
                        :key="pk"
                      >
                        <a
                          v-if="rep.response && rep.response.fileId"
                          style="color: #1470cc; margin-right: 16px;"
                          :href="urlFile + '?key=' + rep.response.fileId"
                          :download="rep.response.fileName"
                        >{{ rep.response.fileName }}</a>
                        <a
                          v-else-if="rep.url"
                          style="color: #1470cc; margin-right: 16px;"
                          :href="rep.url"
                          :download="rep.fileName"
                        >{{ rep.fileName }}</a>
                      </span>
                    </span>
                    <!-- <span v-else-if="mappingKey(k).indexOf('附件') > -1">
                      <span
                        v-for="(rep, pk) in JSON.parse(v)"
                        :key="pk"
                      >
                        <a
                          v-if="rep.response && rep.response.fileId"
                          style="color: #1470cc; margin-right: 16px;"
                          :href="urlFile + '?fileId=' + rep.response.fileId"
                          :download="rep.response.fileName"
                        >{{ rep.response.fileName }}</a>
                      </span>
                    </span>
                    <span v-else-if="['事件报告'].indexOf(mappingKey(k)) > -1">
                      <span
                        v-for="(rep, pk) in JSON.parse(v)"
                        :key="pk"
                      >
                        <a
                          v-if="rep.response && rep.response.fileId"
                          style="color: #1470cc; margin-right: 16px;"
                          :href="urlFile + '?fileId=' +  rep.response && rep.response.fileId"
                          :download=" rep.response && rep.response.fileName"
                        >{{ rep.response && rep.response.fileName }}</a>
                      </span>
                    </span> -->
                    <!-- <span v-else-if="mappingKey(k) === '维修结果'">
                      <span class="log-item-content">
                        {{ '【维修原因】：' + formatResult(v)[0] }}
                        <br>
                        {{ '【维修措施】：' + formatResult(v)[1] }}
                        <br>
                        {{ '【维修结果】：' + formatResult(v)[2] }}
                      </span>
                    </span> -->
                    <span
                      v-else
                      class="text"
                    >{{ v }}</span>
                  </el-form-item>
                </div>
              </div>
              <div v-else>
                {{ item.columns }}
              </div>
            </slot>
          </el-form>
          <el-form
            v-if="$slots.default"
            label-position="right"
            :label-width="'160px'"
          >
            <div>
              <slot />
            </div>
          </el-form>
        </el-collapse-item>
      </el-collapse>
      <div
        v-else
        class="el-table__empty-block"
      >
        <span class="el-table__empty-text">暂无数据</span>
      </div>
    </el-block>
  </div>
</template>

<script>
export default {
  name: 'ElProcessLog',
  props: {
    logs: {
      type: Array,
      required: false,
      default: () => [],
    },
    title: {
      type: String,
      required: false,
      default: '',
    },
    urlImage: {
      type: String,
      required: false,
      default: '',
    },
    urlFile: {
      type: String,
      required: false,
      default: '/cgi/filestorage/downloadFile',
    },
    keyMap: {
      type: Object,
      required: false,
      default: null,
    },
  },
  data() {
    return {
      // phoneIcons: {
      //   拨打中: 'phone-blue2.png',
      //   成功: 'phone-green.png',
      //   未拨通: 'phone-red.png',
      //   未拨打: 'phone-gray.png',
      // },
    };
  },
  computed: {

  },
  methods: {
    reverseIndex(index) {
      return this.logs.length - index;
    },
    // srcUrl(status) {
    //   return `/static/images/icon/${this.phoneIcons[status]}`;
    // },
    formatResult(v) {
      const result = [];
      v.split('\n').map((item) => {
        result.push(item.substr(item.indexOf('：') + 1));
      });
      return result;
    },
    mappingKey(k) {
      if (!this.keyMap) {
        return k;
      }
      return this.keyMap[k] || '';
    },
  },
};
</script>

<style scoped>
@import url('./process-log.css');
.text {
  word-break: break-all;
}
</style>
