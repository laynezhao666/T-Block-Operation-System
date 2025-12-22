
<template>
  <div
    ref="model"
    class="model-container"
  >
    <!-- 模型加载进度条 -->
    <div
      class="loading"
      :style="{ display: progress === 100 ? 'none' : '' }"
    >
      <el-progress
        type="circle"
        :percentage="progress"
        status="text"
        :show-text="true"
        :width="100"
      />
    </div>
    <div
      id="hotpot"
      class="annotation-wrapper"
    />
    <div class="tools-container">
      <div
        class="toolbar-unit"
        @click="toggleCamera('front')"
      >
        <div class="front-icon" />
      </div>
      <div
        class="toolbar-unit"
        @click="toggleCamera()"
      >
        <div class="rotate-icon" />
      </div>
      <div
        class="toolbar-unit"
        @click="screenShot"
      >
        <div class="screenshot-icon" />
      </div>
    </div>
  </div>
</template>

<script>
import * as THREE from 'three';
import {
  sRGBEncoding,
  PMREMGenerator,
  Box3,
  Vector3,
  LinearToneMapping,
  HemisphereLight,
  AmbientLight,
  DirectionalLight,
  // LinearEncoding,
} from 'three';
import Stats from 'three/examples/jsm/libs/stats.module.js';
import { RoomEnvironment } from 'three/examples/jsm/environments/RoomEnvironment.js';
import { OrbitControls } from 'three/examples/jsm/controls/OrbitControls';
import { GLTFLoader } from 'three/examples/jsm/loaders/GLTFLoader.js';
// import { KTX2Loader } from 'three/examples/jsm/loaders/KTX2Loader.js';
// import { DRACOLoader } from 'three/examples/jsm/loaders/DRACOLoader.js';
// import { MeshoptDecoder } from 'three/examples/jsm/libs/meshopt_decoder.module.js';
import { RGBELoader } from 'three/examples/jsm/loaders/RGBELoader.js';
// import { EffectComposer } from 'three/examples/jsm/postprocessing/EffectComposer.js';
// import { RenderPass } from 'three/examples/jsm/postprocessing/RenderPass.js';
// import { OutlinePass } from 'three/examples/jsm/postprocessing/OutlinePass.js';
// import { ShaderPass } from 'three/examples/jsm/postprocessing/ShaderPass.js';
// import { FXAAShader } from 'three/examples/jsm/shaders/FXAAShader.js';
import { TWEEN } from 'three/examples/jsm/libs/tween.module.min';
import { CSS2DRenderer, CSS2DObject } from 'three/examples/jsm/renderers/CSS2DRenderer.js';
import { GUI } from 'three/examples/jsm/libs/lil-gui.module.min.js';

import {
  traverseMaterials,
  environments,
  bloomComponents,
  nameMap,
  traverseObjectParents,
  updateHotpotPosistion,
  hoverComponents,
  computedFrontCamPosition,
  getObjectPosition,
} from '../utils/index.js';

export default {
  components: {},
  props: {
    modelUrl: {
      type: String,
      default: '',
    },
    data: {
      type: Object,
      default: () => ({}),
    },
    ip: {
      type: String,
      default: '',
    },
    isSNMP: {
      type: Boolean,
      default: true,
    },
  },
  data() {
    return {
      activeType: 'front',
      scene: null,
      camera: null,
      model: null,
      renderer: null,
      labelRenderer: null,
      width: 1,
      height: 1,
      observer: null,
      state: {
        environment: environments[1].name,
        background: false,
        // Lights
        punctualLights: true,
        exposure: 0.0,
        toneMapping: LinearToneMapping,
        textureEncoding: 'sRGB',
        ambientIntensity: 0.3,
        ambientColor: 0xffffff,
        directIntensity: 0.8 * Math.PI, // TODO(#116)
        directColor: 0xffffff,
        bgColor1: '#ffffff',
        bgColor2: '#353535',
      },
      lights: [],
      progress: 0,
      raycaster: null,
      composer: null,
      outlinePass: null,
      renderPass: null,
      // DEBUG
      debug: window.location.hash === '#debug',
      sceneMeshes: [],
      bloomObjects: [],
      storeMaterials: {},
      originMaterials: {},
      lightMaterials: [
        new THREE.MeshBasicMaterial({ color: '#00ff00' }),
        new THREE.MeshBasicMaterial({ color: '#ff0000' }),
      ],
      lightMaterial: new THREE.MeshBasicMaterial({ color: '#00ff00' }),
      alarmLightMaterial: new THREE.MeshBasicMaterial({ color: '#ff0000' }),
      cameraPosition: {},
      tweenIds: {},
    };
  },
  watch: {
    data: {
      handler() {
        if (this.content) {
          this.updateLightStatus();
        }
      },
      immediate: true,
      deep: true,
    },
    isSNMP: {
      handler(v) {
        if (this.scene) {
          this.scene.traverse((node) => {
            if (!node.isMesh || node.isCamera || node.isLight) return;
            node.material = v ? new THREE.MeshBasicMaterial({ color: '#cccccc', opacity: 0.5, wireframe: true }) : this.originMaterials[node.uuid];
          });
        }
      },
    },
  },
  mounted() {
    this.initDomSize(this.$refs.model);
    if (this.width && this.height) {
      this.initScene();
    }
    this.$nextTick(() => {
      this.observer = new ResizeObserver(() => {
        this.updateSize();
      });
      this.observer.observe(this.$refs.model);
    });
  },
  beforeDestroy() {
    if (this.observer) {
      this.observer.disconnect();
      this.observer = null;
    }
    this.clear();
  },

  methods: {
    initDomSize(dom) {
      const { width, height, left, top } = dom.getBoundingClientRect();
      this.width = width;
      this.height = height;
      this.left = left;
      this.top = top;
    },
    updateSize() {
      if (!this.width && !this.scene) {
        this.initDomSize(this.$refs.model);
        if (this.width && this.height) {
          this.initScene();
        }
      } else {
        this.initDomSize(this.$refs.model);
        const { clientHeight, clientWidth } = this.$refs.model;
        this.camera.aspect = clientWidth / clientHeight;
        this.camera.updateProjectionMatrix();
        this.renderer.setSize(clientWidth, clientHeight);
        this.labelRenderer.setSize(clientWidth, clientHeight);
        if (this.composer) {
          this.composer.setSize(clientWidth, clientHeight);
        }
        // this.bloomComposer.setSize(clientWidth, clientHeight);
        // this.finalComposer.setSize(clientWidth, clientHeight);
      }
    },
    initScene() {
      // 场景
      this.scene = new THREE.Scene();
      // camera设置
      const fov = 40;
      const aspect = this.width / this.height; // 相机默认值
      this.camera = new THREE.PerspectiveCamera(fov, aspect, 0.01, 1000);
      this.camera.position.set(-1000, 1000, 1000);
      this.scene.add(this.camera);

      // 初始化renderer
      this.renderer = new THREE.WebGLRenderer({
        antialias: true,
        alpha: true,
      });
      this.renderer.physicallyCorrectLights = true;
      this.renderer.outputEncoding = sRGBEncoding;
      // this.renderer.setClearColor(0xffffff);
      this.renderer.setPixelRatio(window.devicePixelRatio || 2);
      this.renderer.setSize(this.width, this.height);
      this.renderer.domElement.style.zIndex = '-1';

      this.labelRenderer = new CSS2DRenderer();
      this.labelRenderer.setSize(this.width, this.height);
      this.labelRenderer.domElement.style.position = 'absolute';
      this.labelRenderer.domElement.style.top = '0px';

      this.labelRenderer.domElement.addEventListener('mousemove', this.onHover, false);
      this.labelRenderer.domElement.addEventListener('click', this.onClick, false);

      this.$refs.model.appendChild(this.renderer.domElement);
      this.$refs.model.appendChild(this.labelRenderer.domElement);

      // 设置环境
      this.pmremGenerator = new PMREMGenerator(this.renderer);
      this.pmremGenerator.compileEquirectangularShader();

      this.neutralEnvironment = this.pmremGenerator.fromScene(new RoomEnvironment()).texture;

      // Controls
      this.controls = new OrbitControls(this.camera, this.labelRenderer.domElement);
      this.controls.autoRotate = false;

      // Raycaster
      this.raycaster = new THREE.Raycaster();

      /** **** outline composer ******/
      // // 创建一个EffectComposer（效果组合器）对象，然后在该对象上添加后期处理通道。
      // this.composer = new EffectComposer(this.renderer);
      // // 新建一个场景通道  为了覆盖到原理来的场景上
      // this.renderPass = new RenderPass(this.scene, this.camera);
      // this.composer.addPass(this.renderPass);
      // this.outlinePass = new OutlinePass(new THREE.Vector2(this.width, this.height), this.scene, this.camera);
      // // 物体边缘发光通道
      // this.outlinePass.edgeStrength = 10.0; // 边框的亮度
      // this.outlinePass.edgeGlow = 1;// 光晕[0,1]
      // // this.outlinePass.usePatternTexture = false; // 是否使用父级的材质
      // this.outlinePass.edgeThickness = 50.0; // 边框宽度
      // this.outlinePass.downSampleRatio = 1; // 边框弯曲度
      // this.outlinePass.pulsePeriod = 50; // 呼吸闪烁的速度
      // this.outlinePass.visibleEdgeColor.set(parseInt(0x00ff00)); // 呼吸显示的颜色
      // // this.outlinePass.hiddenEdgeColor = new THREE.Color(0, 0, 0); // 呼吸消失的颜色
      // this.outlinePass.clear = true;
      // this.outlinePass.selectedObjects = [];
      // this.composer.addPass(this.outlinePass);
      // // 自定义的着色器通道 作为参数
      // this.effectFXAA = new ShaderPass(FXAAShader);
      // this.effectFXAA.uniforms.resolution.value.set(1 / this.width, 1 / this.height);
      // this.effectFXAA.renderToScreen = true;
      // // this.composer.addPass(this.effectFXAA);
      /** **** outline composer ******/

      // Loader
      // this.dracoLoader = new DRACOLoader().setDecoderPath('./static/libs/draco/gltf/');
      // this.ktx2Loader = new KTX2Loader();
      // this.ktx2Loader.setTranscoderPath('./static/libs/basis/');
      this.loader = new GLTFLoader();
      // .setMeshoptDecoder(MeshoptDecoder);
      // .setDRACOLoader(this.dracoLoader)
      // .setKTX2Loader(this.ktx2Loader.detectSupport(this.renderer))
      // .setMeshoptDecoder(MeshoptDecoder);

      // TEST
      const geometry = new THREE.BoxGeometry(100, 100, 100);
      const material = new THREE.MeshBasicMaterial({ color: 0xcccccc });
      const cube = new THREE.Mesh(geometry, material);
      cube.name = 'cube';
      // this.scene.add(cube);
      // this.sceneMeshes.push(cube);
      // this.setContent(cube);

      this.load();
      this.initHotpot();

      this.animate = this.animate.bind(this);
      requestAnimationFrame(this.animate);

      if (this.debug) {
        this.showDebug();
      }
    },
    showDebug() {
      // DEBUG
      this.stats = new Stats();
      this.stats.dom.height = '48px';
      this.stats.dom.style.position = 'absolute';
      this.stats.dom.style.top = '25%';
      this.$refs.model.appendChild(this.stats.dom);
      [].forEach.call(this.stats.dom.children, child => (child.style.display = ''));

      const axesHelper = new THREE.AxesHelper(500);
      this.scene.add(axesHelper);
      this.gui = new GUI();

      this.debugFolder = this.gui.addFolder('camera');
      this.positionDebugFolder = this.debugFolder.addFolder('cameraPosition');
      this.positionDebugFolder.add(this.camera.position, 'x').min(-1000)
        .max(1000)
        .step(1);
      this.positionDebugFolder.add(this.camera.position, 'y').min(-1000)
        .max(1000)
        .step(1);
      this.positionDebugFolder.add(this.camera.position, 'z').min(-1000)
        .max(1000)
        .step(1);
      this.targetDebugFolder = this.debugFolder.addFolder('cameraTarget');
      this.targetDebugFolder.add(this.controls.target, 'x').min(-20)
        .max(20)
        .step(1);
      this.targetDebugFolder.add(this.controls.target, 'y').min(-20)
        .max(20)
        .step(1);
      this.targetDebugFolder.add(this.controls.target, 'z').min(-20)
        .max(20)
        .step(1);
      this.debugFolder.add(this.controls, 'enablePan');
    },
    animate(time) {
      requestAnimationFrame(this.animate);
      TWEEN.update();
      this.controls.update();
      if (this.stats) {
        this.stats.update();
      }
      this.time = time * 0.001;
      this.render();
    },
    render() {
      this.renderer.render(this.scene, this.camera);
      this.labelRenderer.render(this.scene, this.camera);
      if (this.composer) {
        this.composer.render();
      }
    },
    // 加载模型
    load() {
      console.time('load time');
      const url = this.modelUrl || './static/models/caijiqi_ceshi.glb';
      this.loader.load(
        url,
        (gltf) => {
          const scene = gltf.scene || gltf.scenes[0];
          this.sceneMeshes = [];
          scene.traverse((child) => {
            if (child.name.includes('caijiqi')) {
              // child.visible = false;
            }
            if (child.isMesh) {
              const position = getObjectPosition(child);
              child.position.copy(position);
              child.geometry.computeBoundingBox();
              child.geometry.center();
            }

            if (child.isMesh && bloomComponents.includes(child.name)) {
              this.storeMaterials[child.uuid] = child.material;

              this.bloomObjects.push(child);
            }
            if (child.name === 'xianshimianban') {
              const helper = new THREE.BoxHelper(child, 0xff0000);
              helper.update();
              // this.scene.add(helper);
            }
          });
          this.sceneMeshes = [...scene.children];
          // this.outlinePass.selectedObjects = [...this.bloomObjects];
          console.timeEnd('load time');
          console.time('set time');
          this.setContent(scene);
          console.timeEnd('set time');
        },
        (xhr) => {
          this.progress = (xhr.loaded / xhr.total) * 100;
        },
        (error) => {
          console.error(error, 'eeeeee');
        }
      );
    },
    setContent(object) {
      this.clear();
      const box = new Box3().setFromObject(object);
      const size = box.getSize(new Vector3()).length();
      const boxSize = box.getSize(new Vector3());
      const center = box.getCenter(new Vector3());

      object.position.x = -center.x;
      // eslint-disable-next-line no-mixed-operators
      object.position.y = -(center.y + boxSize.y * 0);
      object.position.z = -center.z;
      this.scene.add(object);
      this.content = object;

      this.scene.traverse((node) => {
        if (!node.isMesh || node.isCamera || node.isLight) return;
        this.originMaterials[node.uuid] = node.material;
        node.material = this.isSNMP ? new THREE.MeshBasicMaterial({ color: '#cccccc', opacity: 0.5, wireframe: true }) : node.material;
      });

      this.controls.maxDistance = size * 1;
      this.controls.minDistance = size * 0.8;
      this.camera.near = size / 100;
      this.camera.far = size * 100;

      const viewWidth = boxSize.x * 0.8; // 需要根据容器宽高重新设置
      const viewHeight = Math.max(viewWidth / this.camera.aspect, boxSize.y);
      const viewDeep = viewHeight / Math.tan((20 * Math.PI) / 180);
      // this.controls.target = new Vector3(0, boxSize.y * 0.5, 0);

      // this.camera.lookAt(new Vector3(0, 500, 0));
      const cameraPosition = {
        x: 0,
        y: boxSize.y * 0.8,
        z: viewDeep,
      };
      this.cameraPosition.front = cameraPosition;
      this.cameraPosition.back = {
        ...cameraPosition,
        z: -cameraPosition.z,
      };

      new TWEEN.Tween(this.camera.position).to({ x: cameraPosition.x,
        y: cameraPosition.y,
        z: cameraPosition.z }, 1500)
        .easing(TWEEN.Easing.Exponential.InOut)
        .start();
      this.camera.updateProjectionMatrix();
      new TWEEN.Tween(this.controls.target).to({
        y: boxSize.y * 0.2,
      }, 1500)
        .easing(TWEEN.Easing.Exponential.InOut)
        .start();
      this.cameraPosition.target = this.controls.target;
      this.controls.update();

      this.content.traverse((node) => {
        if (node.isLight) {
          this.state.punctualLights = false;
        } else if (node.isMesh) {
          node.material.depthWrite = !node.material.transparent;
        }
      });

      this.updateLights();
      this.updateEnvironment();
      this.updateTextureEncoding();
    },
    clear() {
      if (!this.content) return;
      TWEEN.removeAll();

      this.scene.remove(this.content);

      // dispose geometry
      this.content.traverse((node) => {
        if (!node.isMesh) return;
        node.geometry.dispose();
      });

      // dispose textures
      traverseMaterials(this.content, (material) => {
        for (const key in material) {
          if (key !== 'envMap' && material[key] && material[key].isTexture) {
            material[key].dispose();
          }
        }
      });
    },
    // todo 加载模型动画
    showLoading() {},
    // todo 模型载入动画
    enseInModel() {},
    updateLightStatus() {
      this.bloomObjects.forEach((object) => {
        let name = object.name?.split('_')[1] || '-';
        if (name.includes('COM') && this.data[name]) {
          if (this.data[name].pv) {
            object.material = this.lightMaterials[+this.data[name].pv];
            this.flashAnim(object, true);
          } else {
            object.material = this.storeMaterials[object.uuid];
            this.flashAnim(object, false);
          }
        } else if (this.data[name] && this.data[name].pv === '0') {
          object.material = this.lightMaterial;
          this.flashAnim(object, true);
        } else if (nameMap[name]) {
          name = nameMap[name];
          if (this.data[name]?.pv === '0' || this.data[name]?.pv === true) {
            object.material = this.lightMaterial;
            this.flashAnim(object, true);
          } else {
            object.material = this.storeMaterials[object.uuid];
            this.flashAnim(object, false);
          }
        } else {
          object.material = this.storeMaterials[object.uuid];
          this.flashAnim(object, false);
        }
      });
    },
    flashAnim(object, animate) {
      if (this.tweenIds[object.id]?.length) {
        if (animate) {
          this.tweenIds[object.id][0].start();
          this.tweenIds[object.id][1].start();
        } else {
          object.scale.set(1, 1, 1);
          this.tweenIds[object.id][0].stop();
          this.tweenIds[object.id][1].stop();
        }
        return;
      }
      if (!animate) {
        return;
      }
      const tweenA = new TWEEN.Tween(object.scale).to({ x: 1.4, y: 1.4, z: 1.4 }, 500)
        .easing(TWEEN.Easing.Sinusoidal.InOut)
        .start();
      const tweenB = new TWEEN.Tween(object.scale).to({ x: 0, y: 0, z: 0 }, 500)
        .easing(TWEEN.Easing.Sinusoidal.InOut);
        // .start();
      tweenA.chain(tweenB);
      tweenB.chain(tweenA);
      if (!this.tweenIds[object.id]) {
        this.tweenIds[object.id] = [tweenA, tweenB];
      }
    },
    updateLights() {
      const { state } = this;
      const { lights } = this;
      if (state.punctualLights && !lights.length) {
        this.addLights();
      } else if (!state.punctualLights && lights.length) {
        this.removeLights();
      } else if (lights.length) {
        lights.forEach((light) => {
          this.scene.add(light);
        });
      }
      this.renderer.toneMapping = Number(state.toneMapping);
      this.renderer.toneMappingExposure = Math.pow(2, state.exposure);

      if (lights.length === 2) {
        lights[0].intensity = state.ambientIntensity;
        lights[0].color.setHex(state.ambientColor);
        lights[1].intensity = state.directIntensity;
        lights[1].color.setHex(state.directColor);
      }
    },
    addLights() {
      const { state } = this;

      const hemiLight = new HemisphereLight();
      hemiLight.name = 'hemi_light';
      this.scene.add(hemiLight);
      this.lights.push(hemiLight);

      const light1 = new AmbientLight(state.ambientColor, state.ambientIntensity);
      light1.name = 'ambient_light';
      this.camera.add(light1);

      const light2 = new DirectionalLight(state.directColor, state.directIntensity);
      light2.position.set(0.5, 0, 0.866); // ~60º
      light2.name = 'main_light';
      this.camera.add(light2);

      this.lights.push(light1, light2);
    },
    removeLights() {
      this.lights.forEach(light => light.parent.remove(light));
      this.lights.length = 0;
    },
    updateEnvironment() {
      const environment = environments.filter(entry => entry.name === this.state.environment)[0];
      this.getCubeMapTexture(environment).then(({ envMap }) => {
        this.scene.environment = envMap;
        this.scene.background = this.state.background ? envMap : null;
      });
    },
    getCubeMapTexture(environment) {
      const { id, path } = environment;
      if (id === 'neutral') {
        return Promise.resolve({ envMap: this.neutralEnvironment });
      }
      // none
      if (id === '') {
        return Promise.resolve({ envMap: null });
      }
      return new Promise((resolve, reject) => {
        new RGBELoader().load(
          path,
          (texture) => {
            const envMap = this.pmremGenerator.fromEquirectangular(texture).texture;
            this.pmremGenerator.dispose();

            resolve({ envMap });
          },
          undefined,
          reject
        );
      });
    },

    updateTextureEncoding() {
      const encoding = sRGBEncoding;
      // this.state.textureEncoding === "sRGB" ? sRGBEncoding : LinearEncoding;
      traverseMaterials(this.content, (material) => {
        if (material.map) material.map.encoding = encoding;
        if (material.emissiveMap) material.emissiveMap.encoding = encoding;
        if (material.map || material.emissiveMap) material.needsUpdate = true;
      });
    },
    onClick(event) {
      const intersects = this.getIntersectObject(event);
      if (this.pickedObject) {
        this.pickedObject = undefined;
      }
      if (intersects.length) {
        this.pickedObject = intersects[0].object;
        // this.toggleObjectView(traverseObjectParents(this.pickedObject));
        console.log(traverseObjectParents(this.pickedObject), 'clickObject');
      } else {
      }
    },
    onHover(event) {
      const intersects = this.getIntersectObject(event);
      if (this.pickedObject) {
        this.pickedObject = undefined;
      }
      if (intersects.length && !intersects[0].object.name.includes('caijiqi')) {
        this.labelRenderer.domElement.style.cursor = 'pointer';
        this.pickedObject = intersects[0].object;
        const pickObjectName = traverseObjectParents(this.pickedObject).name;
        if (this.isSNMP) {
          this.hotpot.textContent = '暂无设备模型';
        } else if (pickObjectName === 'ETH3') {
          this.hotpot.textContent = this.ip;
        } else if (hoverComponents[pickObjectName]) {
          this.hotpot.textContent = hoverComponents[pickObjectName];
        } else {
          this.hotpot.textContent = pickObjectName;
        }
        this.annotation.visible = true;
        updateHotpotPosistion({ pickedObject: this.pickedObject, object: this.annotation });
      } else {
        this.annotation.visible = false;
        this.labelRenderer.domElement.style.cursor = 'auto';
      }
    },
    getIntersectObject(event) {
      // 获取对象
      const { raycaster, renderer, camera } = this;
      const mouse = {
        x: (((event.clientX - this.left) / renderer.domElement.clientWidth) * 2) - 1,
        y: (-((event.clientY - this.top) / renderer.domElement.clientHeight) * 2) + 1,
      };
      raycaster.setFromCamera(mouse, camera);
      const intersects = raycaster.intersectObjects(this.sceneMeshes, true);
      return intersects;
    },
    changeMaterial(object) {
      if (object) {
        const material = new THREE.MeshLambertMaterial({
          transparent: 1,
          opacity: 0.8,
          wireframe: true,
        });
        object.material = material;
        object.material.transparent = 1;
        object.material.opacity = 0.8;
        object.material.wireframe = true;
      } else if (this.selectObject) {
        this.selectObject.material.transparent = 1;
        this.selectObject.material.opacity = 1;
        this.selectObject.material.wireframe = false;
        this.selectObject = null;
      }
    },

    initHotpot() {
      this.hotpot = document.getElementById('hotpot');
      this.annotation = new CSS2DObject(this.hotpot);
      this.annotation.visible = false;
      this.scene.add(this.annotation);
    },

    /** * 辅助函数 */
    screenShot() {
      this.render();
      this.renderer.domElement.toBlob((blob) => {
        const a = document.createElement('a');
        document.body.appendChild(a);
        a.style.display = 'none';
        const url = window.URL.createObjectURL(blob);
        a.href = url;
        a.download = `采集器-${this.ip}.png`;
        a.click();
      });
    },
    toggleObjectView(object) {
      console.log(object, 'object');
      const position = object.name === 'caijiqi001' ? this.cameraPosition.front : computedFrontCamPosition(object);
      const { x, y, z } = position;
      const { camera } = this;

      new TWEEN.Tween(camera.position).to({ x, y, z }, 1500)
        .easing(TWEEN.Easing.Exponential.InOut)
        .start();
      camera.lookAt(x, y, z);
      camera.updateProjectionMatrix();
    },

    toggleCamera(item) {
      if (!item) {
        this.activeType = this.activeType === 'front' ? 'back' : 'front';
      } else {
        this.activeType = item;
      }
      const { x, y, z } = this.cameraPosition[this.activeType];
      new TWEEN.Tween(this.camera.position).to({ x, y, z }, 1500)
        .easing(TWEEN.Easing.Exponential.InOut)
        .start();
    },

  },
};
</script>

<style lang="scss" scoped>
@import "~feature/style/reference";
.model-container {
  width: 100%;
  position: absolute;
  top: -132px;
  bottom: -94px;
  .loading {
    position: absolute;
    left: 0;
    top: 0;
    width: 100%;
    height: 100%;
    display: flex;
    align-items: center;
    justify-content: center;
  }
  .tools-container {
    position: absolute;
    right: 0;
    z-index: 99;
    top: 200px;
    display: flex;
    flex-direction: column;
    button {
      margin: 0;
    }
    .toolbar-unit {
      width: 24px;
      height: 24px;
      text-align: center;
      // background-color: white;
      // box-shadow: $box-shadow;
      // border: 1px solid $border-color;
      + .toolbar-unit {
        margin-top: 12px;
      }
      &.active {
        border: 1px solid $primary;
        color: $primary !important;
      }
      .front-icon {
        height: 100%;
        background: url(../../assets/3D.svg) no-repeat;
        background-size: 100% 100%;
        cursor: pointer;
        &:hover {
          background: url(../../assets/3D-active.svg) no-repeat;
          background-size: 100% 100%;
        }
      }
      .rotate-icon {
        height: 100%;
        background: url(../../assets/rotate.svg) no-repeat;
        background-size: 100% 100%;
        cursor: pointer;
        &:hover {
          background: url(../../assets/rotate-active.svg) no-repeat;
          background-size: 100% 100%;
        }
      }
      .screenshot-icon {
        height: 100%;
        background: url(../../assets/screenshot.svg) no-repeat;
        background-size: 100% 100%;
        cursor: pointer;
        &:hover {
          background: url(../../assets/screenshot-active.svg) no-repeat;
          background-size: 100% 100%;
        }
      }

    }
  }
  .flex-wrap {
    align-items: start;
  }
  .grid {
    display: grid;
    grid-template-columns: 1fr 1fr 1fr;
    .info-item {
      padding: 5px 0 15px 0;
      font-family: TencentSansW3;
      .label {
        width: 80px;
        text-align: right;
        display: inline-block;
        margin-right: 12px;
        color: #bfbcbc;
        font-weight: 800;
      }
      .value {
        color: #333;
      }
    }
  }
  .text-container {
    position: absolute;
    width: 100%;
  }
  .annotation-wrapper {
    background: rgb(255, 255, 255);
    border-radius: 4px;
    box-shadow: rgba(0, 0, 0, 25%) 0px 2px 4px;
    color: rgba(0, 0, 0, 0.8);
    // display: block;
    font-family: Futura, "Helvetica Neue", sans-serif;
    font-size: 18px;
    font-weight: 700;
    max-width: 128px;
    overflow-wrap: break-word;
    padding: 0.5em 1em;
    width: max-content;
    // display: none;
  }
}
</style>
