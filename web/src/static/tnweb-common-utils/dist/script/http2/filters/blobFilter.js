export default function (rsp) {
  if (rsp?.data?.type === 'application/json') {
    const fs = new FileReader()
    fs.readAsText(rsp.data)
    return {
      then (resolve, reject) {
        fs.onload = () => {
          try {
            const rst = JSON.parse(fs.result)
            resolve(rst)
          } catch (e) {
            reject(new Error('响应数据格式错误'))
          }
        }
      },
    }
  }
  return {
    isBlob: true,
    ...rsp,
  }
}
