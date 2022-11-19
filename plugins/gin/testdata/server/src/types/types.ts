export type Params = Param[]

export type GoodsCreateRes = {
  guid?: string;
  selfRef?: SelfRefType;
  Status?: Params;
  stringAlias?: string;
  raw?: string;
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

export type GoodsInfoRes = {
  price?: number;
  title?: string;
  subTitle?: string;
  cover?: string;
}

export type Image = {
  url: string;
}