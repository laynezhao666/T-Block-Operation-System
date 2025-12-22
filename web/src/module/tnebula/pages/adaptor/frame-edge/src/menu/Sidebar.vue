<template>
  <div :class="['custom-sidebar', {'custom-sidebar--collapsed': isCollapsed}]">
    <div class="custom-sidebar-nav">
      <el-menu
        ref="menu"
        unique-opened
        :default-active="activeIndex"
        :collapse="isCollapsed"
        @select="menuCilck"
      >
        <template v-for="item in menuList">
          <el-submenu
            v-if="item.children&&Object.keys(item.children).length>=1"
            :key="item.n_href"
            :index="item.n_href"
          >
            <template slot="title">
              <i
                v-if="item.n_licls && item.n_licls.indexOf('tn-icon') === 0"
                :class="item.n_licls"
              />
              <tn-icon
                v-else-if="item.n_licls"
                :icon="item.n_licls"
              />
              <span slot="title">{{ item.n_name }}</span>
            </template>
            <el-menu-item
              v-for="child in item.children"
              :key="child.n_href"
              :index="child.n_href"
            >
              <span slot="title">{{ child.n_name }}</span>
            </el-menu-item>
          </el-submenu>
          <el-menu-item
            v-else
            :key="item.n_href"
            :index="item.n_href"
          >
            <i
              v-if="item.n_licls && item.n_licls.indexOf('tn-icon') === 0"
              :class="item.n_licls"
            />
            <tn-icon
              v-else-if="item.n_licls"
              :icon="item.n_licls"
            />
            <span slot="title">{{ item.n_name }}</span>
          </el-menu-item>
        </template>
      </el-menu>
    </div>
    <div class="custom-sidebar-footer">
      <slot name="footer" />
      <i
        class="custom-sidebar-footer__collapse-icon tn-icon-arrow-left"
        @click="click"
      />
    </div>
  </div>
</template>

<script>

export default {
  name: 'Sidebar',
  props: {
    activeIndex: {
      default: '',
      type: String,
    },
    collapsed: {
      default: false,
      type: Boolean,
    },
    menuList: {
      default: () => [],
      type: Array,
    },
  },

  data() {
    return {
      isCollapsed: this.collapsed,
      submenuOpened: '',
      activeId: '',
      needOpenMenu: '',
    };
  },

  methods: {
    menuCilck(index) {
      this.$emit('onMenuSelect', index);
    },
    click() {
      this.isCollapsed = !this.isCollapsed;
    },
  },
};
</script>

<style lang="scss" >
.custom-sidebar {
  background: #fff;
  width: 164px;
  height: 100%;
  box-shadow: 3px 3px 8px 2px rgba(216, 216, 216, .5);
  border-right: 1px solid #f0f0f0;
  transition: all .1s ease;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  flex-shrink: 0
}

.custom-sidebar .el-menu-item > span {
  display: inline-block;
  transition: all .1s
}
.custom-sidebar .el-submenu__title{
  width: 164px;
}

.custom-sidebar--collapsed {
  width: 64px;
  overflow: hidden
}

.custom-sidebar-nav .el-submenu.is-active .el-submenu__title :nth-child(1){
  color: #1470cc !important;
}

.custom-sidebar-nav .el-submenu .el-submenu__title>i,
.custom-sidebar-nav .el-menu-item>i{
  margin-right: 16px;

}
.custom-sidebar-nav{
  .el-menu{
    .el-submenu {
      .el-submenu__title{
       padding-left: 16px !important;
      .el-submenu__icon-arrow{
        right: 0px;
      }
    }
     .el-menu--inline .el-menu-item{
     padding-left:56px !important;}
    }
    .el-menu-item{
      padding-left: 16px !important;
    }
  }
  .el-menu--collapse.el-menu{
    .el-menu-item{
      padding-left: 16px !important;
    }

    .el-submenu .el-submenu__title {
      padding-left: 20px !important;
  }
  }

}

// .custom-sidebar-nav .el-submenu .el-submenu__title > .el-submenu__icon-arrow{
//   right: 12px;
// }

// .custom-sidebar-nav .el-menu .el-menu-item,
// .custom-sidebar-nav .el-submenu .el-submenu__title{
//   padding-left:16px !important;
// }
// .el-menu .el-menu--inline .el-menu-item{
//      padding-left:56px !important;
//  }

// .custom-sidebar--collapsed .el-menu-item >span,
// .custom-sidebar--collapsed .el-submenu__title >span,
// .custom-sidebar--collapsed .custom-sidebar-header__text {
//   opacity: 0;
//   visibility: hidden
// }

.custom-sidebar--collapsed .custom-sidebar-header__logo:after {
  opacity: 1
}

.custom-sidebar--collapsed .custom-sidebar-footer__collapse-icon {
  margin-right: 32px;
  transform: rotate(180deg)
}

.custom-sidebar-header {
  box-sizing: border-box;
  height: 72px;
  display: flex;
  flex-shrink: 0;
  align-items: center;
  padding: 0 24px;
  width: 264px
}

.custom-sidebar-header__image {
  width: 40px;
  height: 40px;
  margin-right: 16px
}

.custom-sidebar-header__text {
  transition: all .3s ease
}

.custom-sidebar-nav {
  flex: 1;
  width: 264px;
  border-top: 1px solid #f0f0f0;
  padding: 8px 0;
  overflow-y: auto
}

.custom-sidebar-footer {
  height: 48px;
  border-top: 1px solid #f0f0f0;
  display: flex;
  justify-content: flex-end;
  align-items: center
}

.custom-sidebar-footer__collapse-icon {
  color: #666;
  font-size: 24px;
  margin-right: 16px;
  transition: all .3s ease;
  cursor: pointer
}

.custom-sidebar-footer__collapse-icon.tn-icon-arrow-left {
  margin-right: 16px;
  transition: all .2s;
  color: #333
}

.custom-sidebar-footer__collapse-icon:hover {
  transition: all .2s;
  color: #333
}

</style>
