import { IBaseCurd } from "../base-curd";
import { IBaseTableContext } from "../table-layout-context";

type ITableContextWithCurdAndFormModal = IBaseCurd & {
  curd: IBaseCurd & {
    formModal: ICurdFormModal;
  },
};

export interface ICurdFormModal {
  /** 弹窗标题，默认自动拼接“新增”或“编辑”，传入函数可自定义 */
  title: string | ((isCreate: boolean, edittingRow: any) => string);
  /** 表单内容，需要暴露validate方法用于提交时校验 */
  formComp?: Vue.Component;
  /** 分步骤表单，每个组件comp需要暴露validate方法用于下一步或提交时校验 */
  steps?: {
    /** 步骤标题 */
    title: string;
    comp: Vue.Component;
  }[];
  /** 表单弹窗宽度 */
  width?: number;
  /** 如果需要编辑的表单对象与行对象有格式上的差异，使用该函数进行修改来适配 */
  beforeEdit?: (editting: any, replaceEditting: (newEditting: any) => void) => void;
  submit: (tableContext: ITableContextWithCurdAndFormModal) => void;
  cancel: (tableContext: ITableContextWithCurdAndFormModal) => void;
  onSubmit: (row: any, isCreate: boolean) => Promise<boolean> | boolean;
}

export type ICurdFormModalOptions = Pick<ICurdFormModal, 'title' | 'steps' | 'formComp' | 'onSubmit' | 'beforeEdit'>;

export function curdFormModal<T extends IBaseTableContext>(data: T, opts: ICurdFormModalOptions): T & {
  curd: IBaseCurd & {
    formModal: ICurdFormModal;
  };
} {
  const dataWithCurd: T & { curd: IBaseCurd } = data as unknown as T & { curd: IBaseCurd };

  const { curd } = dataWithCurd;

  const modals = [...data.modals];

  modals.push((() => import('./form-modal.vue')) as any);

  return {
    ...dataWithCurd,
    modals,
    curd: {
      ...curd,
      formModal: {
        ...opts,
        submit,
        cancel,
      }
    }
  };
}

const submit: ICurdFormModal['submit'] = async (tableContext) => {
  const {
    curd: {
      editting,
      isCreate,
      add,
      edit,
      formModal,
    },
  } = tableContext;

  const isSuccess = await formModal.onSubmit(editting, isCreate ||false);

  if (!isSuccess) return;

  close(tableContext);
};

const cancel: ICurdFormModal['cancel'] = (tableContext) => {
  close(tableContext);
};

const close = (tableContext: ITableContextWithCurdAndFormModal) => {
  tableContext.curd.editting = false;
  tableContext.curd.isCreate = null;
};
