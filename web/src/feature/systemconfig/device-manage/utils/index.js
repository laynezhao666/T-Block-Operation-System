

import {
  Box3,
  Vector3,
} from 'three';

export function traverseMaterials(object, callback) {
  object.traverse((node) => {
    if (!node.isMesh) return;
    const materials = Array.isArray(node.material)
      ? node.material
      : [node.material];
    materials.forEach(callback);
  });
}

export const environments = [
  {
    id: '',
    name: 'None',
    path: null,
  },
  {
    id: 'neutral', // THREE.RoomEnvironment
    name: 'Neutral',
    path: null,
  },
  {
    id: 'venice-sunset',
    name: 'Venice Sunset',
    path: 'assets/environment/venice_sunset_1k.hdr',
    format: '.hdr',
  },
  {
    id: 'footprint-court',
    name: 'Footprint Court (HDR Labs)',
    path: 'assets/environment/footprint_court_2k.hdr',
    format: '.hdr',
  },
];

export const modelComponents = [
  // 背面
  'COM_1',
  'COM_2',
  'COM_3',
  'COM_4',
  'COM_5',
  'COM_6',
  'COM_7',
  'COM_8',
  'COM_9',
  'COM_10',
  'DI1',
  'DI2',
  'DI3',
  'DI4',
  'DI5',
  'DI6',
  'DI7',
  'DI8',
  'DI9',
  'DI10',
  'DO1',
  'DO2',
  'DO3',
  'DO4',

  'ETH_1',
  'ETH_2',
  'ETH_3',
  'ETH_4',
  'ETH_5',
  'ETH_6',
  'ETH_7',
  'ETH_8',
  'ETH_9',
  'ETH_10',
  'GND',

  // 正面
  // 'caijiqi.001', // 主体
  'ERR',
  'anniu_01',
  'anniu_02',
  'anniu_03',
  'anniu_04',
  'Console',
  'HDMI',
  'LED_4G',
  'LED_BT',
  'LED_COM1',
  'LED_COM2',
  'LED_COM3',
  'LED_COM4',
  'LED_COM5',
  'LED_COM6',
  'LED_COM7',
  'LED_COM8',
  'LED_COM9',
  'LED_COM10',
  'LED_DI1',
  'LED_DI2',
  'LED_DI3',
  'LED_DI4',
  'LED_DI5',
  'LED_DI6',
  'LED_DI7',
  'LED_DI8',
  'LED_DI9',
  'LED_DI10',
  'LED_DO1',
  'LED_DO2',
  'LED_DO3',
  'LED_DO4',
  'LED_ETH1_2',
  'LED_ETH2_2',
  'LED_ETH3_2',
  'LED_ETH4_2',
  'LED_Main',
  'LED_Online',
  'LED_PWR2',
  'LED_PWR2',
  'LED_RUN',
  'LED_WIFI',
  'Power1',
  'Power2',
  'Rseet',
  'SIM',
  'STA AP',
  'TF',
  'USB.001',
  'xianshimianban',

];

export const hoverComponents = {
  xianshimianban: '显示板',
  Rseet: '重置',
  USB001: 'USB',
};

export const clickComponents = ['xianshimianban', 'USB', 'LED'];

export const bloomComponents = [
  'LED_4G_2',
  'LED_BT_2',
  'LED_COM1_2',
  'LED_COM2_2',
  'LED_COM3_2',
  'LED_COM4_2',
  'LED_COM5_2',
  'LED_COM6_2',
  'LED_COM7_2',
  'LED_COM8_2',
  'LED_COM9_2',
  'LED_COM10_2',
  'LED_DI1_2',
  'LED_DI2_2',
  'LED_DI3_2',
  'LED_DI4_2',
  'LED_DI5_2',
  'LED_DI6_2',
  'LED_DI7_2',
  'LED_DI8_2',
  'LED_DI9_2',
  'LED_DI10_2',
  'LED_DO1_2',
  'LED_DO2_2',
  'LED_DO3_2',
  'LED_DO4_2',
  'LED_ETH1_2',
  'LED_ETH2_2',
  'LED_ETH3_2',
  'LED_ETH4_2',
  'LED_Main_2',
  'LED_Online_2',
  'LED_PWR1_2',
  'LED_PWR2_2',
  'LED_RUN_2',
  'LED_WIFI_2',
  'ERR_2',
];
export const nameMap = {
  PWR1: 'PowerFault_1',
  PWR2: 'PowerFault_2',
  RUN: 'online',
  Online: 'online',
};

export const updateHotpotPosistion = ({ pickedObject, object }) => {
  if (pickedObject === null) return;
  const position = getObjectPosition(pickedObject);
  const { x, y, z } = position;
  object.position.set(x, y + 10, z);
  object.updateMatrixWorld();
};

export const getObjectPosition = (object) => {
  const box = new Box3().setFromObject(object);
  const center = box.getCenter(new Vector3());
  return center;
};

export const getObjectSize = (object) => {
  const box = new Box3().setFromObject(object);
  const size = box.getSize(new Vector3()).length();
  const boxSize = box.getSize(new Vector3());
  return { size, boxSize };
};

export const updateNormal = ({ normal, object }) => {
  if (normal === null) return;
  const { x, y, z } = normal;
  object.normal.set(x, y, z);
};

export const getVertices = (content) => {
  const objects = 0; // 场景模型对象
  let vertices = 0; // 模型顶点
  let triangles = 0; // 模型面片

  content.traverseVisible((object) => {
    if (object.isMesh) {
      const { geometry } = object;
      if (geometry.isGeometry) {
        vertices += geometry.vertices.length;
        triangles += geometry.faces.length;
      } else if (geometry.isBufferGeometry && geometry.attributes.position) {
        vertices += geometry.attributes.position.count;
        if (geometry.index !== null) {
          triangles += geometry.index.count / 3;
        } else {
          triangles += geometry.attributes.position.count / 3;
        }
      }
    }
  });

  console.log(
    `模型对象数量: ${objects}`,
    `模型顶点数: ${vertices}`,
    `模型面片数: ${triangles}`
  );
};

export const traverseObjectParents = (object) => {
  if (object.parent && object.parent.name !== 'Scene') {
    return traverseObjectParents(object.parent);
  }
  return object;
};

export const computedFrontCamPosition = (object) => {
  const position = getObjectPosition(object);
  const { size } = getObjectSize(object);
  const distance = size * 2;
  const camPosition = {
    ...position,
    z: position.z > 0 ? position.z + distance : position.z - distance,
  };
  return camPosition;
};
