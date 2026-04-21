<template>
  <el-form
    ref="form"
    :model="editting"
    :rules="rules"
    label-width="120px"
  >
    <el-form-item
      label="照片"
      prop="picture"
    >
      <el-upload
        class="avatar-uploader"
        action="#" 
        :show-file-list="false"
        :auto-upload="false"
        :before-upload="beforeAvatarUpload"
        :on-change="handleAvatarChange"
      >
        <img v-if="editting.picture" :src="editting.picture" class="avatar">
        <i v-else class="el-icon-plus avatar-uploader-icon"></i>
      </el-upload>
    </el-form-item>

    <el-form-item
      label="姓名"
      prop="name"
      required
    >
      <el-input
        v-model="editting.name"
        placeholder="请输入人员姓名"
      />
    </el-form-item>

    <el-form-item
      label="人员组"
      prop="company"
    >
      <el-input
        v-model="editting.company"
        placeholder="请输入人员组"
      />
    </el-form-item>

    <el-form-item
      label="证件"
      prop="paper"
    >
      <div class="id-card-info-container">
        <el-select
          v-model.trim="editting.paper_type"
          placeholder="证件类型"
          :default-first-option="true"
          class="id-card-type-input"
        >
          <el-option
            label="身份证"
            value="身份证"
          />
        </el-select>

        <el-input
          v-model.trim="editting.paper"
          placeholder="请输入证件号"
          class="id-card-number"
        />
      </div>
    </el-form-item>

    <el-form-item
      label="手机号"
      prop="phone"
    >
      <el-input
        v-model.trim="editting.phone"
        placeholder="请输入手机号"
        type="tel"
      />
    </el-form-item>

    <el-form-item
      label="邮箱"
      prop="email"
    >
      <el-input
        v-model.trim="editting.email"
        placeholder="请输入邮箱"
        type="email"
      />
    </el-form-item>

    <el-form-item
      label="密码"
      prop="password"
    >
      <el-input
        v-model.trim="editting.password"
        placeholder="请输入密码"
        type="password"
        show-password
      />
    </el-form-item>

    <el-form-item
      label="密码确认"
      prop="passwordConfirm"
    >
      <el-input
        v-model.trim="editting.passwordConfirm"
        placeholder="请再次输入密码"
        type="password"
        show-password
      />
    </el-form-item>

    <el-form-item
      label="备注"
      prop="comment"
    >
      <el-input
        v-model.trim="editting.comment"
        placeholder="请输入备注"
        type="textarea"
        :autosize="{ minRows: 2 }"
      />
    </el-form-item>
  </el-form>
</template>

<script>
const ruleCantEmpty = name => ({
  required: true, message: `${name}不能为空`,
});

const idCardRegexp = /(^[1-9]\d{5}(18|19|([23]\d))\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$)|(^[1-9]\d{5}\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}$)/;

export default {
  props: {
    editting: {
      type: Object,
      required: true,
    },
    isCreate: {
      type: Boolean,
      required: true,
    },
  },
  data() {
    return {
      item: [],

      rules: {
        picture: [
            { 
            validator: (rule, value, callback) => {
              // 头像为可选项，如果没有上传则通过验证
              if (!this.editting.picture) {
                callback();
                return;
              }
              
              // 如果上传了图片，检查格式是否以 data:image 开头的 Base64 格式
              if (!this.editting.picture.startsWith('data:image')) {
                callback(new Error('请上传有效的头像图片'));
              } else {
                callback();
              }
            }
          }
        ],
        name: [
          ruleCantEmpty('姓名'),
        ],
        paper: [
          {
            validator(rule, value, cb) {
              const { paper, oldPaper } = this;
              if (paper === oldPaper) return cb();

              return idCardRegexp.text(paper);
            },
            message: '证件号不正确',
          },
        ],
        phone: [
          { pattern: /^1[0-9]{10}$/, message: '手机号格式不正确' },
        ],
        email: [
          { type: 'email', message: '邮箱格式不正确' },
        ],
        password: [
          {
            pattern: /^\d{6}$|^\*{6}$/,
            message: '请输入6位数字密码',
          },
        ],
        passwordConfirm: [
          {
            validator: (rule, value, callback) => {
              const isSameWithPassword = (value || '') === (this.editting.password || '');
              callback(isSameWithPassword ? undefined : '两次输入的密码不同，请检查后重新输入。');
            },
          },
        ],
      },
    };
  },
  methods: {
    beforeAvatarUpload(file) {
      // 获取文件扩展名（小写）
      const extension = file.name.split('.').pop().toLowerCase();
      // 允许的图片扩展名
      const allowedExtensions = ['jpg', 'jpeg', 'png'];

      const isImage = allowedExtensions.includes(extension);
      const isLt300KB = file.size / 1024 < 300;

      if (!isImage) {
        this.$message.error('上传头像只能是 JPG/PNG 格式!');
        return false;
      }
      if (!isLt300KB) {
        this.$message.error('上传头像大小不能超过 300KB!');
        return false;
      }
      
      return true;
    },
    
    // 检查图片分辨率
    checkImageResolution(file) {
      return new Promise((resolve, reject) => {
        const img = new Image();
        const objectUrl = URL.createObjectURL(file.raw);
        
        img.onload = () => {
          URL.revokeObjectURL(objectUrl); // 释放内存
          const maxWidth = 720;
          const maxHeight = 1280;
          
          if (img.width > maxWidth || img.height > maxHeight) {
            reject(new Error(`图片分辨率不能超过 ${maxWidth}×${maxHeight}，当前分辨率为 ${img.width}×${img.height}`));
          } else {
            resolve(true);
          }
        };
        
        img.onerror = () => {
          URL.revokeObjectURL(objectUrl);
          reject(new Error('图片加载失败'));
        };
        
        img.src = objectUrl;
      });
    },
    
    async handleAvatarChange(file) {
      // 先进行基础验证（格式和大小）
      console.log('开始验证图片');

      const isValid = this.beforeAvatarUpload(file);
      if (!isValid) return false;
      console.log('验证图片格式和大小成功');

      // 检查图片分辨率
      try {
        await this.checkImageResolution(file);
        console.log('验证图片分辨率成功');
      } catch (error) {
        this.$message.error(error.message);
        return false;
      }

      // 使用FileReader将图片转为Base64
      const reader = new FileReader();
      reader.onload = (e) => {
        this.$set(this.editting, 'picture', e.target.result); // 使用$set确保响应式更新
        this.$refs.form.validateField('picture'); // 手动触发头像字段验证
      };
      reader.onerror = () => {
        this.$message.error('图片读取失败，请重试');
      };
      reader.readAsDataURL(file.raw);
      
      return false;
    },

    async validate() {
      try {
        // 手动触发头像字段验证
        await new Promise((resolve) => {
          this.$refs.form.validateField('picture', (error) => {
            resolve(!error);
          });
        });
        
        // 验证整个表单
        await this.$refs.form.validate();
        
        // 如果是编辑模式且没有修改头像，保留原头像
        if (!this.isCreate && !this.editting.picture) {
          this.editting.picture = this.editting.oldPicture;
        }
        return true;
      } catch (error) {
        return false;
      }
    }
  },
};
</script>

<style lang="scss" scoped>
.id-card-info-container {
  display: flex;
  align-items: center;
}

.id-card-type-input {
  width: 120px;
  margin-right: 8px;
}

.avatar-uploader {
  ::v-deep .el-upload {
    border: 1px dashed #d9d9d9;
    border-radius: 6px;
    cursor: pointer;
    position: relative;
    overflow: hidden;
    
    &:hover {
      border-color: #409EFF;
    }
  }
  
  .avatar-uploader-icon {
    font-size: 28px;
    color: #8c939d;
    width: 178px;
    height: 178px;
    line-height: 178px;
    text-align: center;
  }
  
  .avatar {
    width: 178px;
    height: 178px;
    display: block;
    object-fit: cover;
  }
}

</style>
