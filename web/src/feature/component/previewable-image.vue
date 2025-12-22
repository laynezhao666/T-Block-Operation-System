<template>
  <span
    v-loading="status === 'loading'"
    class="previewable-img"
    @click="handleClick"
  >
    <el-image
      v-show="src"
      ref="img"
      :src="src"
      :fit="fit"
      :lazy="lazy"
      :alt="alt"
      :scroll-container="scrollContainer"
      :class="status === 'loaded' ? 'img' : ''"
      @load="handleLoad"
      @error="handleError"
    />
    <span
      v-show="!src"
    >
      --
    </span>
  </span>
</template>

<script>
export default {
  props: [
    'src',
    'fit',
    'lazy',
    'alt',
    'scrollContainer',
    'previewStyle',
    'width',
    'height',
  ],
  data() {
    return {
      status: 'loading',
    };
  },
  watch: {
    src() {
      this.status = 'loading';
    },
    // '$refs.img': {
    //   handler: function ($img) {
    //     if (!$img) {
    //       return
    //     }
    //     $img.$el.addEventListener('click', () => this.handleClick())
    //   },
    //   immediate: true,
    // },
  },
  // mounted () {
  //   this.$refs.img.$el.addEventListener('click', () => this.handleClick())
  // },
  methods: {
    handleLoad(e) {
      this.$emit('load', e);
      this.status = 'loaded';
    },
    handleError(e) {
      this.$emit('error', e);
      this.status = 'error';
    },
    handleClick() {
      if (this.status !== 'loaded') {
        return;
      }

      const {
        src,
        alt,
        previewStyle,
      } = this;
      this.$alert(`<img src="${src}" class="previewable-image-img" alt="${alt}" style="${previewStyle}" />`, alt, {
        dangerouslyUseHTMLString: true,
      });
    },
  },
};
</script>

<style lang="scss" scoped>
.previewable-img {
  cursor: pointer;
  position: relative;
  display: inline-block;

  &:hover {
    &::before {
      content: '';
      display: block;
      position: absolute;
      top: 0;
      left: 0;
      width: 100%;
      height: 100%;
      background-color: rgba(0,0,0,0.5);
      z-index: 2;
    }

    &::after {
      content: '\ee9a6';
      z-index: 2;
      color: #fff;

      position: absolute;
      top: 50%;
      left: 50%;
      top: calc(50% - 8px);
      left: calc(50% - 8px);

      font-family: 'tnweb-icons' !important;
      font-style: normal;
      font-weight: normal;
      font-variant: normal;
      text-transform: none;
      line-height: 1;
      -webkit-font-smoothing: antialiased;
    }
  }
}
</style>

<style>
.previewable-image-img {
  display: block;
  margin: auto;
}
.el-message-box {
  width: auto !important;
  min-width: 384px;
}
</style>
