/*
 * @description is a single URL parameter, consisting of a key and a value.
 */
export type Param = {
  Key?: string;
  Value?: string;
}

/*
 * @description is a Param-slice, as returned by the router.
 *	The slice is ordered, the first URL parameter is also the first slice value.
 *	It is therefore safe to read values by the index.
 */
export type Params = Param[]

export enum ErrCode {
  CodeNotFound = 10000,
  CodeCancled = 10001,
  CodeUnknown = 10002,
  CodeInvalidArgument = 10003,
}

export type Error = {
  code?: ErrCode;
  msg?: string;
}

export type GoodsCreateReq = {
  /*
   * @description 封面图
   */
  cover?: string;
  /*
   * @description 详情图
   */
  images?: Image[];
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

export type GoodsCreateRes = {
  /*
   * @description 测试引用第三方包
   */
  Status?: Params;
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
  selfRef?: SelfRefType;
  /*
   * @description 测试类型别名
   */
  stringAlias?: string;
}

export type GoodsDownRes = {
  Status?: string;
}

export type GoodsInfoRes = {
  cover?: string;
  mapInt?: Record<number, Property>;
  price?: number;
  properties?: Record<string, Property>;
  subTitle?: string;
  title?: string;
}

/*
 * @description 商品图片
 */
export type Image = {
  /*
   * @description 图片链接
   */
  url: string;
}

export type Property = {
  title?: string;
}

export type SelfRefType = {
  data?: string;
  parent?: SelfRefType;
}