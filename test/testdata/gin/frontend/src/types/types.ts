/*
 * @description
 */
export type CustomResponseType<A> = {
  /*
   * @description
   */
  code: number;
  /*
   * @description
   */
  data: A;
  /*
   * @description
   */
  msg: string;
}

/*
 * @description
 */
export type ShopGoodsDownRequest = {
  dateRange?: string[];
  defaultPostForm?: string;
  operatorUid?: string;
}

export type GinParam = {
  /*
   * @description
   */
  Key?: string;
  /*
   * @description
   */
  Value?: string;
}

export type GinParams = GinParam[]

/*
 * @description
 */
export type GormDeletedAt = string

/*
 * @description
 */
export enum ViewErrCode {
  CodeNotFound = 10000,
  CodeCancled = 10001,
  CodeUnknown = 10002,
  CodeInvalidArgument = 10003,
}

/*
 * @description
 */
export type ViewError = {
  /*
   * @description
   */
  code?: ViewErrCode;
  /*
   * @description
   */
  msg?: string;
}

/*
 * @description
 */
export type ViewGoodsCreateReq = {
  cover?: string;
  images?: ViewImage[];
  price: number;
  subTitle?: string;
  title: string;
}

/*
 * @description
 */
export type ViewGoodsCreateRes = {
  Status?: GinParams;
  guid?: string;
  raw?: any;
  selfRef?: ViewSelfRefType;
  stringAlias?: string;
}

/*
 * @description
 */
export type ViewGoodsDownRes = {
  /*
   * @description
   */
  Status?: string;
}

/*
 * @description
 */
export type ViewGoodsInfoRes = {
  /*
   * @description
   */
  cover?: string;
  /*
   * @description
   */
  deletedAt?: GormDeletedAt;
  /*
   * @description
   */
  mapInt?: Record<number, ViewProperty>;
  /*
   * @description
   */
  price?: number;
  /*
   * @description
   */
  properties?: Record<string, ViewProperty>;
  /*
   * @description
   */
  subTitle?: string;
  /*
   * @description
   */
  title?: string;
}

export type ViewImage = {
  url: string;
}

/*
 * @description
 */
export type ViewProperty = {
  /*
   * @description
   */
  title?: string;
}

/*
 * @description
 */
export type ViewSelfRefType = {
  /*
   * @description
   */
  data?: string;
  /*
   * @description
   */
  parent?: ViewSelfRefType;
}

