// import { ENV_NAME } from 'common/script/passport_login';
import http from 'common/script/http2';
import { bigData } from './cgi';

export const getTopolchargedQuery = (
  moduleId,
  graphId,
  isTedge = false
) => {
  const params = {
    mozuID: moduleId,
  };

  if (graphId) {
    params.graphId = graphId;
  }

  // eslint-disable-next-line max-len
  const url = isTedge ? bigData.getTopolchargedQueryEdge : bigData.getTopolchargedQuery;

  return http.post(`${url}?mozuID=${moduleId}`, params, false, {
    isJson: true,
    restAxios: {
      headers: {
        mozuId: moduleId,
        platform: 'cloud',
      },
    },
  });
};

export default {};
