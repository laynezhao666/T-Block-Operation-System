# 表格布局&上下文
封装常用的表格布局，可配置表格、远程或本地分页、批量操作等。

## 基础使用
共三个步骤：
1. 引入布局组件
2. 创建上下文，传入获取表格数据方法
3. 给布局传入上下文

```:javascript
<template>
  <tedge-table-layout
    :context="tableLayoutContext"
  >
    <template #columns>
      <el-table-column
        prop="name"
        label="tbox名称"
      />
    <template>
  </tedge-table-layout>
</template>
<script>
import TedgeTableLayout from 'feature/component/tedge-components/tedge-table-layout.vue';
import { chainTableLayout } from 'feature/component/tedge-components/table-layout-context/table-layout-context';

export default {
  components: {
    TedgeTableLayout,
  },
  data() {
    return {
      tableLayoutContext: this.createTableLayoutContext(),
    };
  },
  methods: {
    createTableLayoutContext() {
      return chainTableLayout(this.fetchData.bind(this));
    },
    async fetchData(filters, search, pagination) {
      return ...;
    },
  },
};
</script>
```

## slots
- columns: 列插槽，同el-table的default插槽，主要放el-table-column
  - 列是不同需求里变化概率最大，故而不做过度封装，用更贴近原生的方式
- toolbar-extra: 工具栏的额外参数，由于工具栏右侧
- outer-modals: 外部弹窗，针对需要自定义弹窗的场景，可将弹窗写于该slot内

## 上下文方法
>> chainTableLayout含ts定义，只是可能不是非常准

### 供调用方法
- ctx.loadData()
  - 重新加载数据
- ctx.forceReloadData()
  - 强制重新加载数据
  - 适用一些缓存了fetchData结果的场景，如本地分页

### 配置方法
- ctx.pagination(opts?: { current: number size: number, total: number, pageSizes: number })
  - 添加分页器功能，需结合本地或远程过滤分页使用
- ctx.localFilterPagination()
  - 开启本地分页，自动分页
  - 注意：由于分页缓存了fetchData的数据，对于需要刷新表格需要调用ctx.forceReloadData()方法，强制从服务器重新拉取数据
- ctx.remoteFilterPagination(opts?: {...})
  - 远程筛选分页，将会给ctx.fetchData传入filter和search、pagination
  - opts.totalFields: 总数字段名
  - opts.listFields: 列表字段名
- ctx.selection({ identity: (row => string | number) | string, oprs: ['export' | 'delete' | VueComponent], hideToolbar: 隐藏工具栏 })
  - 开启批量选择操作，将会每行显示复选框，选择了行以后会展示多选操作工具栏
  - identify: 字段字符串，或返回行标识的函数，用于标记所选的行
  - oprs: 批量操作功能，数组，内置功能可用字符串，其余需自行实现Vue组件传入；内置方法有导出'export'、批量删除'delete'，批量删除会调用上下文中curd模块的批量删除功能。
  - hideToolbar: 是否隐藏工具栏
- ctx.radioRowSelect: 单选行选择器（由于使用场景取消，未做测试）
- ctx.tableStyle(opts)
  - 设置表格样式
  - opts为el-table可用参数，详见组件库
  - 常用：border、stripe、size、fit、showHeader、highlightCurrentRow、height、rowClassName、spanMethod
- ctx.indexColumn(opts?: {...})
  - 开启序号/索引列，在第一列显示序号列
  - opts.label 标题
  - opts.width 宽度
  - opts.fixed 固定位置，'left' | 'right'
- ctx.filters(filtersData: Record<string, any>, opts?: {...})
  - 开启过滤
  - filtersData: 传入过滤数据，ctx内部直接引用、不拷贝、调用方可直接修改该数据
  - opts.isResetPagination: 布尔值，默认true，过滤条件变化是否自动重置分页器
  - opts.filtersForm: 过滤表单组件
- ctx.extraBtn(Comp: VueComponent)
  - 添加工具栏按钮
  - 多次调用可添加多个按钮
- ctx.toolbarActions(opt)
  - 添加工具栏下拉操作，详见el-table-toolbar组件的actions参数
  - opt.text: 操作文案
  - opt.icon: 操作图标
  - opt.action: 操作函数
- ctx.search(opts?: {...})
  - 启用关键字搜索，即el-table-toolbar的search部分
  - opts.value: 初始搜索内容
  - opts.placeholder: 占位符
  - opts.isHide: 隐藏
  - opts.doSearch: 搜索操作函数，((ctx) => any)，覆盖默认搜索行为，一般不用穿
- ctx.hideToolbar()
  - 隐藏工具栏

CURD配置
- ctx.baseCurd(opts: {...})
  - 启用基础curd配置，新增、编辑需与表单配置结合使用
  - 内置add、edit、remove操作，add位于工具栏，其他两个位于每行的最后一列
  - 可自定义行操作，opts.rowOprsComponents
  - OprConfig: 操作配置类型
    - 类型签名：{label: string, action: (row) => boolean, disabled: (row) => boolean,adminRight: boolean }
    - label: 展示文案，如按钮文案
    - action: 函数，点击触发事件，返回false表示停止，返回非false值表示作为新的行设置到editting里，如果返回true则认为返回了空对象{}
    - disabled: 函数，返回是否禁用
    - adminRight: 布尔值，是否限定管理员权限
  - opts.add: 新增配置,类型OprConfig，配置null、undefine则不显示新增按钮
  - opts.edit: 编辑配置,类型OprConfig，配置null、undefine则不显示新增按钮
  - opts.remove: 删除配置,类型OprConfig & {confirm: {}, batchRemove: Function}，配置null、undefine则不显示新增按钮
    - opts.remove.confirm: 布尔值或字符串，是否显示确认气泡，若是字符串则气泡提示内容为字符串
    - opts.remove.remove: 删除回调，参数row
    - opts.remove.batchRemove?: 批量删除回调，参数rows
  - opts.rowEditColumnWidth: 操作列宽
  - opts.rowOprsComponents: 自定义列操作，VueComponent数组
    - 组件参数：row行数据，index行号，table-context表格上下文

CURD表单弹窗
- ctx.curdFormModal(opts: {...})
  - 配置curd使用的表单弹窗
  - opts.title: 字符串或函数，弹窗标题，若是字符串则自动拼接“新增”或“编辑”，传入函数可自定义((isCreate: boolean, edittingRow: any) => string)
  - opts.formComp: 表单组件，VueCompnent
    - 组件参数：editting、isCreate
    - 组件方法：validate，表单验证
  - opts.steps: 分步骤表单，不同步骤同样编辑一个edittingRow
    - opts.steps.title: 步骤标题
    - opts.steps.comp: 同opts.formComp
  - opts.width: 弹窗宽度
  - opts.beforeEdit: 行编辑前前置处理函数，参数editting: any, replaceEditting: (newEditting: any) => void
  - opts.onSubmit: 提交前处理，签名((row: any, isCreate: boolean) => Promise<boolean> | boolean)

## 工具函数
### 创建按钮方法
使用：
```::javascript
import { curryingRenderElTextButton } from 'feature/component/tedge-components/table-layout-context/render-el-text-button';

const btn = curryingRenderElTextButton({
  adminRight: true,
  label: '重命名',
  onClick: (props) => {
    this.$prompt('新名称', '重命名')
      .then(async ({ value: newName }) => {
        await axiosPut('/api/dcos/tsim-cgi/collector/name', {
          id: props.row.id,
          name: newName,
        });
        this.tableLayoutContext.forceReloadData();
      }).catch(() => {});
  },
});
```

- curryingRenderElTextButton(opts)，创建文本按钮
  - opts.adminRigth: boolean，是否管理员权限
  - opts.label: string或返回字符串函数，按钮文案
  - opts.labelComp: 优先级高于label，自定义label组件，参数见上下文
  - opts.extraComps: 渲染额外组件，参数见上下文，如tableContext、row、rows，如需按钮触发打开弹窗可考虑将弹窗置于该参数、并设置弹窗的append-to-body参数，并在onClick里通过extraComps参数访问，
  - opts.onClick(props, extraComps): 点击回调，参数props看上下文，在表格curd上下文中，一般会携带row行数据
  - opts.confirm: 是否启用确认，在执行onClick方法前谈气泡询问用户是否确认操作
    - opts.confirm.title: 确认标题内容
  - opts.disabled: 禁用方法回调，返回布尔值或者字符串，若是字符串则悬浮时提示禁用原因
  - opts.btnProps: 按钮参数，参考el-button参数
