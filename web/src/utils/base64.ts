export const fileToDataBase64 = (file: File) => {
  return new Promise((resolve, reject) => {
    let reader = new FileReader();
    reader.onload = function (evt) {
      let base64 = evt.target?.result;
      resolve(base64);
    };
    reader.readAsDataURL(file);
  });
};