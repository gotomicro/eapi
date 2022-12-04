export type Image = {
  url: string;
}

export type GoodsCreateReq = {
  subTitle?: string;
  cover?: string;
  price: number;
  images?: Image[];
  title: string;
}

export type Error = {
  code?: string;
}

export type Params = Param[]

export type GoodsCreateRes = {
  stringAlias?: string;
  raw?: any;
  guid?: string;
  selfRef?: SelfRefType;
  Status?: Params;
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

export type GoodsDownRes = {
  Status?: string;
}

export type SelfRefType = {
  data?: string;
  parent?: SelfRefType;
}

export type Param = {
  Key?: string;
  Value?: string;
}