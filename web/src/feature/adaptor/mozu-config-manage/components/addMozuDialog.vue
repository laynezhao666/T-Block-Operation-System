<template>
  <el-modal :visible.sync="dialogVisible" title="新增模组" :width="800">
    <el-form ref="form" :model="formData" label-width="80px">

      <el-form-item label-width="120px" label="园区" prop="belong_campus" :rules="{
        required: true, message: '不能为空', trigger: 'blur'
      }">
        <el-input v-model="formData.belong_campus" style="width: 360px;"></el-input>
      </el-form-item>

      <el-form-item label-width="120px" label="园区编码" prop="belong_campus_code" :rules="{
        required: true, message: '不能为空', trigger: 'blur'
      }">
        <el-input v-model="formData.belong_campus_code" style="width: 360px;"></el-input>
      </el-form-item>
      <el-form-item label-width="120px" label="楼栋" prop="belong_building" :rules="{
        required: true, message: '不能为空', trigger: 'blur'
      }">
        <el-input v-model="formData.belong_building" style="width: 360px;"></el-input>
      </el-form-item>

      <el-form-item label-width="120px" label="模组名称" prop="mozu_name" :rules="{
        required: true, message: '名称不能为空', trigger: 'blur'
      }">
        <el-input v-model="formData.mozu_name" style="width: 360px;"></el-input>
      </el-form-item>

      <el-form-item label-width="120px" label="模组编码" prop="mozu_code" :rules="{
        required: true, message: '不能为空', trigger: 'blur'
      }">
        <el-input v-model="formData.mozu_code" style="width: 360px;"></el-input>
      </el-form-item>

      <el-form-item label-width="120px" label="模组ID" prop="mozu_id" :rules="{
        required: true, message: 'ID不能为空', trigger: 'blur'
      }">
        <el-input v-model="formData.mozu_id" style="width: 360px;"></el-input>
      </el-form-item>

      <el-form-item label-width="120px" label="版本" prop="version" :rules="{
        required: true, message: '版本不能为空', trigger: 'blur'
      }">
        <el-input v-model="formData.version" style="width: 360px;"></el-input>
      </el-form-item>
      <el-form-item label-width="120px" v-for="item in formItems" :label="item.label" :prop="item.value"
        :key="item.value" :rules="{ required: true, message: '不能为空', trigger: 'blur' }">
        <el-upload action="" style="width: 360px;" :http-request="(file) => customUpload(file, item.value)"
          :on-preview="handlePreview" :on-remove="handleRemove" :before-remove="beforeRemove" multiple :limit="3"
          :on-exceed="handleExceed" :file-list="fileList">
          <el-button size="small"><i class="tn-icon-upload"></i>上传</el-button>
        </el-upload>
      </el-form-item>
    </el-form>
    <template #footer>
      <el-button @click="dialogVisible = false">取消</el-button>
      <el-button type="primary" @click="submitForm">保存</el-button>
    </template>
  </el-modal>
</template>

<script>
import { tbosImportCgi } from '@@/config/cgi.js'
import { omit } from 'lodash';

export default {
  components: {
  },
  props: {
    visible: {
      type: Boolean,
      default: false,
    },
  },
  data() {
    return {
      formData: {
        mozu_id: '',
        mozu_name: '',
        version: '',
        belong_building:"",
        belong_campus: "",
        belong_campus_code: "",
        mozu_code: "",
        "alarm_strategy": {
          "content": null,
          "filename": ""
        },
        "collector_device": {
          "content": null,
          "filename": ""
        },
        "collector_template": {
          "content": null,
          "filename": ""
        },
        "device_entity": {
          "content": null,
          "filename": ""
        },
        "device_point": {
          "content": null,
          "filename": ""
        },
        "template_point": {
          "content": null,
          "filename": ""
        },
      },
      formItems: [
        {
          label: '采集模板',
          value: 'collector_template',
        },
        {
          label: '采集设备',
          value: 'collector_device',
        },
        {
          label: '设备实体',
          value: 'device_entity',
        },
        {
          label: '设备测点',
          value: 'device_point',
        },
        {
          label: '测点模板',
          value: 'template_point',
        },
        {
          label: '告警策略',
          value: 'alarm_strategy',
        },
      ],
      fileList: [],
      mozuListOptions: [],
      mozuOptionsMap: {},
      // 级联选择器配置
      cascaderProps: {
        value: 'id',    // 指定 value 字段
        label: 'name',  // 指定 label 字段
        emitPath: false,
      },
    };
  },
  computed: {
    dialogVisible: {
      set(v) {
        this.$emit('update:visible', v);
      },
      get() {
        if (!this.visible) {
          this.formData = {
            mozu_id: '',
            mozu_name: '',
            version: '',
            "alarm_strategy": {
              "content": null,
              "filename": ""
            },
            "collector_device": {
              "content": null,
              "filename": ""
            },
            "collector_template": {
              "content": null,
              "filename": ""
            },
            "device_entity": {
              "content": null,
              "filename": ""
            },
            "device_point": {
              "content": null,
              "filename": ""
            },
            "template_point": {
              "content": null,
              "filename": ""
            },
            "version": ""
          }
        }
        return this.visible;
      },
    },
  },
  created() {
  },
  methods: {
    submitForm() {
      this.$refs.form.validate((valid) => {
        if (valid) {
          this.handleMozuSave()
        }
      });
    },
    handleRemove(file, fileList) {
      console.log(file, fileList);
    },
    handlePreview(file) {
      console.log(file);
    },
    handleExceed(files, fileList) {
      this.$message.warning(`当前限制选择 3 个文件，本次选择了 ${files.length} 个文件，共选择了 ${files.length + fileList.length} 个文件`);
    },
    beforeRemove(file, fileList) {
      return this.$confirm(`确定移除 ${file.name}？`);
    },
    async getMozuList() {
      const { moduleGroups } = await this.$axios.post(tbosImportCgi.getMozu, {})
      const processedGroups = moduleGroups.map((group) => {
        // 第一级ID处理：id+片区
        const level1 = {
          ...group,
          id: `${group.id}+片区`
        }
        // 如果有子级，处理第二级ID
        if (level1.children && level1.children.length) {
          level1.children = level1.children.map((child) => {
            // 处理第三级数据，生成mozuOptionsMap
            if (child.children && child.children.length) {
              child.children.forEach((thirdLevel) => {
                this.mozuOptionsMap[thirdLevel.id] = thirdLevel.name
              })
            }

            return {
              ...child,
              id: `${child.id}+园区`
            }
          })
        }

        return level1
      })
      this.mozuListOptions = processedGroups
    },
    handleMozuChange() {

    },
    async handleMozuSave() {
      console.log(this.mozuOptionsMap)
      const isAllUploaded = this.formItems.every(item => {
        return this.formData[item.value] && this.formData[item.value].content
      })
      // if (!isAllUploaded) {
      //   this.$message.warning('请上传所有文件');
      //   return;
      // }
      await this.$axios.post(tbosImportCgi.saveMozu, {
        "mozu_id": this.formData.mozu_id,
        "mozu_name": this.formData.mozu_name,
        belong_building: this.formData.belong_building,
        belong_campus: this.formData.belong_campus,
        belong_campus_code: this.formData.belong_campus_code,
        mozu_code: this.formData.mozu_code,
      })

      // 转换formData中的二进制内容为Base64
      const formDataWithBase64 = {
        ...omit(this.formData, ['mozu_name', 'belong_building', 'belong_campus', 'belong_campus_code', 'mozu_code'])
      }

      this.formItems.forEach(item => {
        if (formDataWithBase64[item.value].content) {
          // 将ArrayBuffer转换为Base64
          const uint8Array = new Uint8Array(formDataWithBase64[item.value].content);
          let binary = '';
          uint8Array.forEach(byte => {
            binary += String.fromCharCode(byte);
          });
          formDataWithBase64[item.value].content = btoa(binary);
        }
      });

      await this.$axios.post(tbosImportCgi.importModel, {
        ...formDataWithBase64
      })
      this.$emit('confirm');
    },
    customUpload(file, v) {
      console.log(file);
      console.log(v);
      return new Promise((resolve) => {
        const reader = new FileReader()
        reader.readAsArrayBuffer(file.file)
        reader.onload = (e) => {
          const binaryData = e.target.result;
          console.log('文件名称:', file.file.name);
          console.log('文件内容:', binaryData);
          this.formData[v] = {
            filename: file.file.name,
            content: binaryData
          }
          resolve(); // 模拟上传成功
        }
      });
    }
  },
};
</script>

<style lang="scss" scoped></style>