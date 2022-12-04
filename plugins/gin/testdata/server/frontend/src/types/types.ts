export type SelfRefType = {
  parent?: SelfRefType;
  data?: string;
}

export type Param = {
  Key?: string;
  Value?: string;
}

export type GoodsDownRes = {
  Status?: string;
}

export type GoodsCreateRes = {
  raw?: any;
  guid?: string;
  selfRef?: SelfRefType;
  Status?: Params;
  stringAlias?: string;
}

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
  title: string;
  subTitle?: string;
  cover?: string;
  price: number;
  images?: Image[];
}

export type Error = {
  code?: string;
}

export type Params = Param[]