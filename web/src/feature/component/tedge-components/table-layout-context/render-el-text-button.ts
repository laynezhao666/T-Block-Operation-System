import AdminLimitLoginAlert from 'feature/component/tedge-components/admin-limit-login-alert.vue';

export interface ICurryingRenderElTextButtonOptions {
  label: string | ((props: { [key: string]: any }) => string);
  labelComp?: Vue.Component,
  extraComps?: Vue.Component[],
  confirm?: {
    title: string;
  };
  disabled?: (props: { [key: string]: any }) => boolean | string;
  onClick: (props: { [key: string]: any }, vm?: any) => void;

  adminRight?: boolean,

  btnProps?: {
    [key: string]: any;
  } | ((props: { [key: string]: any }) => {
    [key: string]: any;
  });
}

export const curryingRenderElTextButton = (opts: ICurryingRenderElTextButtonOptions): Vue.Component => ({
  functional: true,
  render(h, ctx) {
    const {
      confirm,
      disabled,
    } = opts;

    const handleClick = () => {
      opts.onClick(ctx.props, extraCompsInstances.map(vnode => vnode.componentInstance));
    }

    const extraCompsInstances = (opts.extraComps || []).map(comp => h(comp, {
      props: { ...ctx.props },
      ref: 'extraComps',
      refInFor: true,
    }));

    const createBtn = (isDisabled = false) => h(
      'el-button', {
      props: {
        type: 'text',
        disabled: isDisabled,
        ...(typeof opts.btnProps === 'function' ? opts.btnProps(ctx.props) : (opts.btnProps || {})),
      },
      on: {
        ...(confirm ? {} : { click: handleClick }),
      },
    },
      [
        opts.labelComp
          ? h(opts.labelComp, {
            props: {
              ...ctx.props,
            },
          }) : (typeof opts.label === 'function' ? opts.label(ctx.props) : opts.label),
        ...extraCompsInstances,
      ]
    );

    const disabledWithAdminRight = (props: Record<string, any>) => {
      if (opts.adminRight) {
        const hasRight = window.tnwebServices.loginStatusService.hasRight();
        if (!hasRight) return AdminLimitLoginAlert;
      }

      return disabled ? disabled(props) : false;
    };


    const disabledResult = disabledWithAdminRight(ctx.props);
    if (disabledResult) {
      const disabledBtn = createBtn(Boolean(disabledResult));

      if (disabledResult) {
        const disabledResultType = typeof disabledResult;
        if (disabledResultType === 'string') {
          return h('el-tooltip', {
            props: {
              content: disabledResult,
            },
          }, [h('span', null, [disabledBtn])]);
        } else if (disabledResultType !== 'boolean') {
          return h('el-tooltip', {
            // scopedSlots: {
            //   content: () => 'button',
            // },
          }, [
            h('span', {
              slot: 'default'
            }, [disabledBtn]),
            h(disabledResult, {
              props: {
                effect: 'dark',
              },
              slot: 'content',
            }),
          ]);
        }
      }
    }

    if (confirm) {
      return h('el-popconfirm', {
        props: {
          ...confirm,
        },
        on: {
          onConfirm: handleClick,
        },
        scopedSlots: {
          reference: () => createBtn(),
        },
      });
    }

    return createBtn();
  },
});
