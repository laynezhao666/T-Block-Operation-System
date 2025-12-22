// 位置定义
import { getMozuId } from 'feature/utils/business';

const currentMozuId = getMozuId();
const locationMap = {
  464: {
    it: [3, 4, 1, 2],
    elec: [6, 7, 10, 11, 0, 1, 4, 5, 9, 3, 2, 8],
  },
  386: {
    it: [1, 2, 3, 4],
    elec: [0, 1, 6, 7, 5, 4, 11, 10, 9, 3, 8, 2],
  },
};
const locationList = locationMap[currentMozuId]?.it || [1, 2, 3, 4];
const locationElecList = locationMap[currentMozuId]?.elec || [0, 1, 6, 7, 4, 5, 10, 11, 3, 9, 2, 8];

export function getLocations() {
  const roomNames = locationList.map(i => `M10${i}`);
  return {
    roomNames,
  };
}
// IT方仓映射
let words = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'U'];
const newArray = [...words].reverse();

const roomMaps = {};
// const roomOrder = [3, 4, 1, 2];
const roomOrder = locationList;
for (let i = 0; i < 4; i++) {
  for (let k = 0; k < words.length; k++) {
    let key = '';
    if ([1, 3].includes(i)) {
      key = `M10${roomOrder[i]}-ITM-${newArray[k]}`;
    } else {
      key = `M10${roomOrder[i]}-ITM-${words[k]}`;
    }
    roomMaps[key] = { room: 'itRooms', index: [i, k] };
  }
}

// 空调方仓映射
words = ['01', '02', '03', '04', '05', '06', '07', '08', '09'];
for (let i = 1; i < 5; i++) {
  for (let k = 0, j = 0; k < words.length; k++) {
    const key = `M10${roomOrder[i - 1]}-IEAC-${words[k]}`;
    roomMaps[key] = { room: 'coolRooms', index: [i - 1, j] };
    j = j + 1;
  }
  i = i + 1;
  // for (let k = words.length - 1, j = 0; k >= 0; k--) {
  for (let k = 0, j = 0; k <= words.length - 1; k++) {
    const key = `M10${roomOrder[i - 1]}-IEAC-${words[k]}`;
    roomMaps[key] = { room: 'coolRooms', index: [i - 1, j] };
    j = j + 1;
  }
}

// 中低压方仓 locationElecList
roomMaps['PAL101-LVPM-01'] = { room: 'electricRooms', index: [locationElecList[0]] };
roomMaps['PAL101-LVPM-02'] = { room: 'electricRooms', index: [locationElecList[1]] };
roomMaps['PAL101-LVPM-03'] = { room: 'electricRooms', index: [locationElecList[2]] };
roomMaps['PAL101-LVPM-04'] = { room: 'electricRooms', index: [locationElecList[3]] };
roomMaps['PAL101-LVPM-05'] = { room: 'electricRooms', index: [locationElecList[4]] };
roomMaps['PAL101-LVPM-06'] = { room: 'electricRooms', index: [locationElecList[5]] };
roomMaps['PAL101-LVPM-07'] = { room: 'electricRooms', index: [locationElecList[6]] };
roomMaps['PAL101-LVPM-08'] = { room: 'electricRooms', index: [locationElecList[7]] };
// 中压方仓
roomMaps['PAL101-HVPM-01'] = { room: 'electricRooms', index: [locationElecList[8]] };
roomMaps['PAL101-HVPM-02'] = { room: 'electricRooms', index: [locationElecList[9]] };
// 并机方仓
roomMaps['PAL101-PSPM-01'] = { room: 'electricRooms', index: [locationElecList[10]] };

export const boxMaps = roomMaps;

const elecNameList = [
  'LVPM-01',
  'LVPM-02',
  'LVPM-03',
  'LVPM-04',
  'LVPM-05',
  'LVPM-06',
  'LVPM-07',
  'LVPM-08',
  'HVPM-01',
  'HVPM-02',
  'PSPM-01',
  '仓库',
];

const roomPlaMap = {};

elecNameList.forEach((i, index) => {
  const key = locationElecList[index];
  roomPlaMap[key] = i;
});

console.log(roomPlaMap, 'roomPlaMaproomPlaMap');

// const roomPlaMap = {
//   0: 'LVPM-05',
//   1: 'LVPM-06',
//   6: 'LVPM-01',
//   7: 'LVPM-02',
//   4: 'LVPM-07',
//   5: 'LVPM-08',
//   10: 'LVPM-03',
//   11: 'LVPM-04',
//   9: 'HVPM-01',
//   3: 'HVPM-02',
//   2: 'PSPM-01',
// };

const itms = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'J', 'K', 'M', 'N', 'P', 'Q', 'R', 'S', 'T', 'U'];

export function getRooms() {
  const itRooms = [];

  for (let i = 0; i < 4; i++) {
    const arr = [];
    for (let j = 0; j < 18; j++) {
      arr.push({
        status: 0,
        name: '',
        address: '',
        no: itms[j],
      });
    }
    itRooms.push(arr);
  }

  const coolRooms = [];
  for (let i = 0; i < 4; i++) {
    const arr = [];
    for (let j = 0; j < 9; j++) {
      arr.push({
        status: 0,
        name: '',
        address: '',
        no: j + 1,
      });
    }
    coolRooms.push(arr);
  }
  const electricRooms = [];

  for (let i = 0; i < 12; i++) {
    electricRooms.push({
      status: 0,
      name: '',
      address: '',
      no: roomPlaMap[i],
    });
  }
  // 通过方仓名，找到位置
  // 4个空调数组，4个it数组，1个低压数组

  const box = {
    itRooms, // 二维，4个房间，每房间18个
    coolRooms, // 二维，4个房间，每房间9个
    electricRooms, // 一维，12个
  };
  return box;
}

export function resetRooms(rooms) {
  for (let i = 0; i < rooms.itRooms.length; i++) {
    const room = rooms.itRooms[i];
    for (let j = 0; j < room.length; j++) {
      room[j].status = 0;
    }
  }
  for (let i = 0; i < rooms.coolRooms.length; i++) {
    const room = rooms.coolRooms[i];
    for (let j = 0; j < room.length; j++) {
      room[j].status = 0;
    }
  }
  for (let i = 0; i < rooms.electricRooms.length; i++) {
    const room = rooms.electricRooms[i];
    room.status = 0;
  }
}
