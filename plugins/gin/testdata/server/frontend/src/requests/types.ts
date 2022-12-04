export type ErrCode = unknown

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
  selfRef?: SelfRefType;
  Status?: Params;
  stringAlias?: string;
  raw?: any;
  guid?: string;
}

export type Property = {
  title?: string;
}

export type GoodsInfoRes = {
  subTitle?: string;
  cover?: string;
  price?: number;
  properties?: Record<string, Property>;
  mapInt?: Record<number, Property>;
  title?: string;
}

export type Image = {
  url: string;
}

export type GoodsCreateReq = {
  images?: Image[];
  title: string;
  subTitle?: string;
  cover?: string;
  price: number;
}

export type Params = Param[]

export type GoodsDownRes = {
  Status?: string;
}