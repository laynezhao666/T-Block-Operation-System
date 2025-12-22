
export const downloadByUrl = (url: string, fileName?: string) => {
  const linkElt = document.createElement('a');
  linkElt.href = url;
  linkElt.style.display = 'none';

  if (fileName) {
    linkElt.download = fileName;
  } else {
    linkElt.target = '_blank';
  }

  document.body.append(linkElt);
  linkElt.click();
  document.body.removeChild(linkElt);
}

export const downloadByObject = (data: Object | Array<any>, fileName: string) => {
  const blob = new Blob([JSON.stringify(data)]);

  const linkElt = document.createElement('a');
  linkElt.href = window.URL.createObjectURL(blob);
  linkElt.style.display = 'none';

  linkElt.download = fileName;

  document.body.append(linkElt);
  linkElt.click();
  document.body.removeChild(linkElt);
}
