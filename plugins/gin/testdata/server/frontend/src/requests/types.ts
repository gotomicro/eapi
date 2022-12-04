export type Param = {
  Key?: string;
  Value?: string;
}

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
  cover?: string;
  images?: Image[];
  price: number;
  subTitle?: string;
  title: string;
}

export type GoodsCreateRes = {
  Status?: Params;
  guid?: string;
  raw?: any;
  selfRef?: SelfRefType;
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

export type Image = {
  url: string;
}

export type Property = {
  title?: string;
}

export type SelfRefType = {
  data?: string;
  parent?: SelfRefType;
}