import * as _ from "lodash";
import { curryingRenderElTextButton } from "../render-el-text-button";
import { IBaseTableContext } from "../table-layout-context";

export interface IBaseCurd {
  editting: null | any;
  isCreate: boolean | null;
  add?: {
    /** 添加按钮文本，默认“新增” */
    label?: string;
    /** 点击触发事件，返回false表示停止，返回非false值表示作为新的行设置到editting里，如果返回true则认为返回了空对象{} */
    action?: () => boolean | any;
    /** 不可用状态函数，返回true/false/string，若为字符串则会提示用户不可用原因 */
    disabled?: (row: any[]) => string;
    /** 是否限定管理员权限 */
    adminRight?: boolean;
  };
  edit?: {
    /** 编辑按钮文本，默认“编辑” */
    label?: string;
    /** 点击触发事件，返回false表示停止，返回非false值表示作为新的行设置到editting里，如果返回true则认为返回了空对象{} */
    action?: (row) => boolean;
    /** 判断行是否可编辑 */
    disabled?: (row) => boolean;
    /** 是否限定管理员权限 */
    adminRight?: boolean;
  };
  remove?: {
    /** 编辑按钮文本，默认“删除” */
    label?: string;
    /** 是否需要确认，默认true */
    confirm?: boolean | string;
    /** 点击触发事件，返回false表示停止，返回非false值表示作为新的行设置到editting里，如果返回true则认为返回了空对象{} */
    action?: (row: any, index: number) => boolean;
    /** 判断行是否可编辑 */
    disabled?: (row) => boolean;
    /** 删除回调 */
    remove?: (rows: any) => Promise<boolean>;
    /** 批量删除回调，与selection等配合实现删除功能 */
    batchRemove?: (rows: any) => Promise<boolean>;
    /** 是否限定管理员权限 */
    adminRight?: boolean;
  };
  rowOprsComponents: Array<Vue.Component>;
  rowEditColumnWidth?: number;
}

export function baseCurd<T extends IBaseTableContext>(data: T, opts?: Partial<IBaseCurd>): T & { curd: IBaseCurd } {
  const curd: IBaseCurd = {
    editting: null,
    isCreate: null,
    add: {},
    edit: {},
    remove: {},
    rowOprsComponents: [],
    ...opts,
  };

  const oprsColumnOprs = [
    ...data.oprsColumnOprs,
    ...(opts?.rowOprsComponents || []),
  ];
  const extras = [...data.extras];

  if (curd.add) {
    extras.push(curryingRenderElTextButton({
      label: curd.add.label || '新增',
      adminRight: curd.add.adminRight,
      disabled: (props) => props.tableContext.curd.add.disabled?.(props.tableContext.tableData) || false,
      btnProps: {
        type: 'primary',
      },
      onClick: async (props: { [key: string]: any; }) => {
        const tableContext: T & { curd: IBaseCurd } = props.tableContext;
        const actionResult = await  curd.add?.action?.() || null;

        if (actionResult === false) return;
        tableContext.curd.editting = actionResult === true || !actionResult ? {} : actionResult;
        tableContext.curd.isCreate = true;
      }
    }));
  }

  if (curd.edit) {
    oprsColumnOprs.push(curryingRenderElTextButton({
      label: curd.edit?.label || '编辑',
      adminRight: curd.edit?.adminRight,
      disabled: curd.edit?.disabled,
      onClick: async (props: { [key: string]: any; }) => {
        const tableContext: T & { curd: IBaseCurd } = props.tableContext;

        let editting = _.cloneDeep(props.row);
        if (curd.edit?.action) {
          editting = await curd.edit.action(editting);
        }

        tableContext.curd.editting = editting;
        tableContext.curd.isCreate = false;
      }
    }));

    // oprsColumnOprs.push({
    //   functional: true,
    //   render: (h, ctx) => h('el-button', {
    //     props: {
    //       type: 'text',
    //       disabled: curd.edit?.disabled?.(ctx.props.row),
    //     },
    //     on: {
    //       click: async () => {
    //         const tableContext: T & { curd: IBaseCurd } = ctx.props.tableContext;

    //         let editting = _.cloneDeep(ctx.props.row);
    //         if (curd.edit?.action) {
    //           editting = await curd.edit.action(editting);
    //         }

    //         tableContext.curd.editting = editting;
    //         tableContext.curd.isCreate = false;
    //       }
    //     },
    //   }, [
    //     curd.edit?.label || '编辑',
    //   ]),
    // });
  }

  if (curd.remove) {
    oprsColumnOprs.push((() => import('./remove-row-btn.vue')) as any);
  }

  return {
    ...data,
    extras,
    oprsColumnOprs,
    curd,
  };
}
