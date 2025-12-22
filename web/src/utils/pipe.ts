
export const pipeWithPromise = (funcList: Function[]) => async (data: any) => {
  let tmp: any = data;
  for(let i = 0, l = funcList.length; i < l; i++) {
    tmp = await funcList[i](tmp);
  }
  return tmp;
};
