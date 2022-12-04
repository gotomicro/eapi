export type GoodsInfoRes = {
  cover?: string;
  price?: number;
  properties?: Record<string, Property>;
  mapInt?: Record<number, Property>;
  title?: string;
  subTitle?: string;
}

export type GoodsCreateReq = {
  images?: Image[];
  title: string;
  subTitle?: string;
  cover?: string;
  price: number;
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
  guid?: string;
  selfRef?: SelfRefType;
  Status?: Params;
  stringAlias?: string;
  raw?: any;
}