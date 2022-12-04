export type Property = {
  title?: string;
}

export type Image = {
  url: string;
}

export type Error = {
  code?: ErrCode;
  msg?: string;
}

export type SelfRefType = {
  data?: string;
  parent?: SelfRefType;
}

export type Param = {
  Key?: string;
  Value?: string;
}

export type GoodsCreateRes = {
  raw?: any;
  guid?: string;
  selfRef?: SelfRefType;
  Status?: Params;
  stringAlias?: string;
}

export type GoodsInfoRes = {
  title?: string;
  subTitle?: string;
  cover?: string;
  price?: number;
  properties?: Record<string, Property>;
  mapInt?: Record<number, Property>;
}

export type GoodsCreateReq = {
  title: string;
  subTitle?: string;
  cover?: string;
  price: number;
  images?: Image[];
}

export enum ErrCode {
  CodeNotFound = 10000,
  CodeCancled = 10001,
  CodeUnknown = 10002,
  CodeInvalidArgument = 10003,
}

export type Params = Param[]

export type GoodsDownRes = {
  Status?: string;
}