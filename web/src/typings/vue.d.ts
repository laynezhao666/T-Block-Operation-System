import Vue from 'vue';

declare module 'vue/types/vue' {
  interface Vue {
    $moduleInfo: ModuleInfo;
  }

  interface ModuleInfo {
    building: string;
    buildingId: number;
    mozu: string;
    mozuId: number;
    mozuNumber: string;
    park: string;
    parkId: number;
  }
}