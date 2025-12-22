declare module '*.svg'
declare module '*.png'
declare module '*.jpg'
declare module '*.jpeg'
declare module '*.gif'
declare module '*.bmp'
declare module '*.tiff'
declare module '*.vue'

type Constructor = new (...args: any[]) => any

declare const ht: {
  Shape: any;
  HistoryManager: new () => HtHistoryManager;
  Default: any;
  Node: typeof HtNode;
  Edge: any;
  Block: any;
  JSONSerializer: any;
  DataModel: typeof HtDataModel;
  List: typeof HtList;

  graph: {
    GraphView: typeof HtGraphView;
    Interactor: typeof HtInteractor
  },
};

declare class HtHistoryManager {
  clear: () => void;
  beginTransaction: () => void;
  endTransaction: () => void;
  setDisabled: (isDisabled: boolean) => void;
  setHistoryIndex: (index: number) => void;
}

declare abstract class HtInteractor {
  constructor(graphView: HtGraphView);

  gv: HtGraphView;

  abstract beginTransaction(): void;
  abstract endTransaction(): void;
  abstract handle_mousedown(evt: any): void;
  abstract handle_touchstart(evt: any): void;
  abstract handleWindowMouseMove(evt: any): void;
  abstract handleWindowMouseUp(evt: any): void;
  abstract handleWindowTouchEnd(evt: any): void;

  addListeners(): void;
  removeListeners(): void;
  getView(): HTMLElement;
  setUp(): void;
  tearDown(): void;
  clear(): void;
  /** graphView.fireInteractorEvent(evt) */
  fi(evt: any): void;
  /** graphView.setCursor(cssCursor) */
  setCursor(cssCursor: string): void;
  startDragging(data: any): void;
  clearDragging(): void;
  autoScroll(data: any): void;
  /** 由交互管理器注入 */
  endInteractor(node: HtNode): void;
}

declare class HtNode{
  dm(): HtDataModel;

  _source?: HtNode;
  _target?: HtNode;
  addChild(childNode: HtNode);
  setIcon(icon: string);
  setRect(rect: IRect);
  getPoints?: () => HtList<IPosition>;
  setPoints(pointList: HtList<any> | any[]);
  /** Edge类型才有 */
  getTarget?: () => HtNode | undefined;
  setTarget(target: HtNode | null | undefined);
  /** Edge类型才有 */
  getSource?: () => HtNode | undefined;
  setSource(target: HtNode | null | undefined);
  /** Edge类型才有 */
  _lastTargetPoint: any;
  /** Edge类型才有 */
  _lastSourcePoint: any;
  // comp?: import('@@/components-manage/base-component').default | null;

  a(attrs: { [key: string]: any }): any;
  a(key: string): any;
  a(key: string, newValue: any): any;
  getAttr(key: string): any;
  getAttrObject(): { [key: string]: any } | undefined;
  setAttrObject(attrObj: { [key: string]: any }): any;

  s(styles: { [key: string]: any }): any;
  s(key: string): any;
  s(key: string, value: any): any;
  getStyleMap(): { [key: string]: any; };
  setStyleMap(styleMap: { [key: string]: any; }): void;

  getId(): number;
  getTag(): string;
  setTag(tag: string);

  setDisplayName(newDisplayName: string);
  getDisplayName(): string;

  getImage(): string | undefined;
  setImage(imageNameOrUrl: string);

  getLayer(): string;
  setLayer(layer: string);

  getRect(): IRect;

  getPosition(): IPosition;
  setPosition(position: IPosition);
  getX(): number;
  setX(x: number);
  getY(): number;
  setY(y: number);

  getWidth(): number;
  setWidth(value: number);

  getHeight(): number;
  setHeight(value: number);

  getTall(): number;
  setTall(value: number);

  getScale(): number;
  setScale(): number;

  getRotation(): number;
  setRotation(rotation: number);

  /** 设置位置挂载对象，一般设置为父节点，host节点移动时，该节点也会跟随移动 */
  getHost: (() => HtNode | undefined) | undefined;
  /** 设置位置挂载对象，一般设置为父节点，host节点移动时，该节点也会跟随移动 */
  setHost: ((node?: HtNode) => void) | undefined;
  /** 获取吸附子节点列表，当该节点移动时，吸附子节点列表也会跟随移动 */
  getAttaches: (() => HtList<HtNode>) | undefined;
  /** 设置吸附子节点列表，当该节点移动时，吸附子节点列表也会跟随移动 */
  setAttaches: ((nodes: Array<HtNode>) => void) | undefined;

  /** 获取连线的边 */
  getEdges(): HtList<HtNode>;

  /** 设置中心点, anchorX和anchorY取值均应在0~1之间，即中心点位置的比例 */
  setAnchor(anchorX: number, anchorY: number, isKeepRect?: boolean);
  /** 设置中心点X轴, 取值均应在0~1之间，即中心点位置的比例 */
  setAnchorX(anchorX: number);
  /** 设置中心点Y轴, 取值均应在0~1之间，即中心点位置的比例 */
  setAnchorY(anchorY: number);
  /** 获取中心点 */
  getAnchor();
  getAnchorX();
  getAnchorY();

  setToolTip(message: string);

  removeFromDataModel();

  /** 获取父级节点 */
  getParent(): HtNode | undefined;
  /** 设置父级节点 */
  setParent(parentNode: HtNode | undefined);
  getChildren(): HtList<HtNode>;
}

declare class HtList<T> {
  addAll(arr: T[]);
  size(): number;
  forEach(func: (node: T) => void): void;
  toArray(): Array<T>;
  getArray(): Array<T>;
}

declare interface IHtEvent {
  kind: string;
  data: HtNode;
}

declare type IHtEventHandler = ((event: IHtEvent) => void);

declare interface IRect {
  x: number;
  y: number;
  width: number;
  height: number;
}

declare class HtGraphView {
  setFocus(e: any): void;
  getLogicalPoint(event: any): IPosition;
  setHeight(height: number): void;
  setWidth(width: number): void;
  /** 视图位置x */
  getTranslateY(): number;
  /** 视图位置y */
  getTranslateX(): number;
  /** 获取缩放，图纸到位到像素单位的缩放 */
  getZoom(): number;

  /** 比例自适应容器大小 */
  getViewRect(): IRect;
  /** 监听事件 */
  mi(func: IHtEventHandler);
  /** 取消监听 */
  umi(func: IHtEventHandler);

  /** 比例自适应容器大小 */
  fitContent();

  /** 允许显示tooltip */
  enableToolTip();
  /** 禁止显示tooltip */
  disableToolTip();

  /** 获取挂在的元素(div) */
  getView(): HTMLDivElement;

  /** 根据位置获取节点 */
  getDataAt(pos: IPosition): HtNode;

  /** 获取dataModel */
  dm(): HtDataModel;

  /** 添加到某个dom中显示 */
  addToDOM(elt: HTMLElement);
}

declare class HtJSONSerializer {
  serialize: () => string;
}

declare class HtDataModel {
  a: HtNode['a'];
  getAttrObject: HtNode['getAttrObject'];
  setAttrObject: HtNode['setAttrObject'];

  moveTo(node: HtNode, index: number);
  moveToIndex(node: HtNode, index: number);
  getRoots(): HtList<HtNode>;
  remove(child: HtNode): void;
  serialize(): string;
  deserialize(str: string): HtList<HtNode>;
  /** 清空图元 */
  clear(): void;
  add(HtNode);
  getDatas(): HtList<HtNode>;
  getDataById(id: number): HtNode | undefined;
  getDataByTag(tag: string): HtNode | undefined;
  sm(): HtSelectionModel;
  getSelectionModel(): HtSelectionModel;

  beginTransaction(): void;
  endTransaction(): void;
}

declare class HtSelectionModel {
  co(node: HtNode): boolean;
  ss(nodes: HtNode[]): void;
  getSelection(): HtList<HtNode>;
  toSelection(): HtList<HtNode>;
}

declare interface IPosition {
  x: number;
  y: number;
}
