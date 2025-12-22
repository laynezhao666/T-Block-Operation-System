export interface ISelectFileOptions {
  multiple: boolean | undefined;
  accept: string | undefined;
}

/** 选择文件函数 */
export const selectFile = async (accept?: string): Promise<File> => {
  return createElementTmp('input', (inputElt: HTMLInputElement) => {
    inputElt.type = 'file';
    inputElt.style.display = 'none';

    if (accept) {
      inputElt.accept = accept;
    }

    inputElt.click();

    return new Promise<File>((resolve, reject) => {
      inputElt.onchange = (evt) => {
        if (!inputElt.files) return;
        resolve(inputElt.files[0]);
      };
    });
  });
}

export const selectAndReadAsTextFile = async (accept?: string): Promise<String> => {
  const file = await selectFile(accept);
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.addEventListener('load', (event) => {
      resolve(reader.result as string);
    });
    reader.readAsText(file);
  });
}

/** 插入dom元素到body，返回移除函数 */
export const appendHiddenEltToBody = (elt: HTMLElement) => {
  document.body.append(elt);

  return () => {
    document.body.removeChild(elt);
  };
}

/** 临时创建dom元素，直到func函数执行完成后删除该dom元素（如果返回promise则等待promise完成或报错） */
export const createElementTmp = <T extends HTMLElement, R>(eltTag: string, func: (elt: T) => R): R => {
  const elt = document.createElement(eltTag) as unknown as T;

  const remove = appendHiddenEltToBody(elt);

  const result = func(elt);

  if (result instanceof Promise) {
    result.finally(remove);
    return result;
  }

  remove();
  return result;
};
