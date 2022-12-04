export type Property = {
  title?: string;
}

export type GoodsInfoRes = {
  title?: string;
  subTitle?: string;
  cover?: string;
  price?: number;
  properties?: Record<string, Property>;
  mapInt?: Record<number, Property>;
}

export type Image = {
  url: string;
}

export type GoodsCreateReq = {
  cover?: string;
  price: number;
  images?: Image[];
  title: string;
  subTitle?: string;
}

export type Params = Param[]

export type GoodsDownRes = {
  Status?: string;
}

export type ErrCode = unknown

export type Error = {
  code?: ErrCode;
  msg?: string;
}

export type SelfRefType = {
  parent?: SelfRefType;
  data?: string;
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