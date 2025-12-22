
import config from '@@/config/business';
import Cookie from 'js-cookie';

export function getMozuId() {
  // return 464;
  const mozu = window.__GetFrameDataByKey('curMozuData');

  if (!Object.prototype.hasOwnProperty.call(mozu, 'id')) {
    return Cookie.get('tnebula_cu_moduleid') && parseInt(Cookie.get('tnebula_cu_moduleid'));
  }
  return (mozu && mozu.id && parseInt(mozu.id));
}

export function getMozuName() {
  const mozu = window.__GetFrameDataByKey('curMozuData');
  return (mozu && mozu.name);
}

if (config.isTedge) {
  const tnbl = window.TNBL || (window.TNBL = {});
  if (!tnbl.getCurrModule) {
    tnbl.getCurrModule = () => {
      const id = getMozuId();
      const name = getMozuName();
      return {
        id,
        name,
      };
    };
  }
}
