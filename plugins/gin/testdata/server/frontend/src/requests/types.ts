export type CustomResponseType<A> = {
  code: number;
  data: A;
  msg: string;
}

export type ShopGoodsDownRequest = {
  /*
   * @description 日期范围
   */
  dateRange?: string[];
  /*
   * @description Default Post Form
   */
  defaultPostForm?: string;
  /*
   * @description 操作人 UID
   */
  operatorUid?: string;
}

/*
 * @description Param is a single URL parameter, consisting of a key and a value.
 */
export type GinParam = {
  Key?: string;
  Value?: string;
}

/*
 * @description Params is a Param-slice, as returned by the router.
 *	The slice is ordered, the first URL parameter is also the first slice value.
 *	It is therefore safe to read values by the index.
 */
export type GinParams = GinParam[]

export type GormDeletedAt = string

export enum ViewErrCode {
  CodeNotFound = 10000,
  CodeCancled = 10001,
  CodeUnknown = 10002,
  CodeInvalidArgument = 10003,
}

export type ViewError = {
  code?: number;
  msg?: string;
}

export type ViewGoodsCreateReq = {
  /*
   * @description 封面图
   */
  cover?: string;
  /*
   * @description 详情图
   */
  images?: ViewImage[];
  /*
   * @description 价格(分)
   */
  price: number;
  /*
   * @description 商品描述
   */
  subTitle?: string;
  /*
   * @description 商品标题
   */
  title: string;
}

export type ViewGoodsCreateRes = {
  /*
   * @description 测试引用第三方包
   */
  Status?: GinParam[];
  /*
   * @description 商品 GUID
   */
  guid?: string;
  /*
   * @description 测试引用内置包类型
   */
  raw?: any;
  /*
   * @description 测试循环引用
   */
  selfRef?: ViewSelfRefType;
  /*
   * @description 测试类型别名
   */
  stringAlias?: string;
}

export type ViewGoodsDownRes = {
  Status?: string;
}

export type ViewGoodsInfoRes = {
  cover?: string;
  deletedAt?: string;
  mapInt?: Record<number, ViewProperty>;
  price?: number;
  properties?: Record<string, ViewProperty>;
  subTitle?: string;
  title?: string;
}

/*
 * @description Image 商品图片
 */
export type ViewImage = {
  /*
   * @description 图片链接
   */
  url: string;
}

export type ViewProperty = {
  title?: string;
}

export type ViewSelfRefType = {
  data?: string;
  parent?: ViewSelfRefType;
}