export type UploaderUploadFileRequest = {
  file?: File;
  /*
   * @description If set true, this file would be saved to lib
   */
  saveToLib?: string;
}

export type ModelArray<T> = T[]

export type ModelCreateGoodsRequest = {
  /*
   * @description Url of cover image
   */
  cover?: string;
  /*
   * @description Detail images
   */
  images?: ModelImage[];
  status?: ModelGoodsStatus;
  stock?: number;
  title?: string;
}

/*
 * @description GenericTypeResponse used for testing generic type
 */
export type ModelGenericTypeResponse<T> = {
  data?: T;
  metadata?: Record<string, any>;
  value?: ModelSampleGenericType<T>;
}

export type ModelGoodsInfo = {
  /*
   * @description Url of cover image
   */
  cover?: string;
  /*
   * @description Unique key
   */
  id?: number;
  /*
   * @description Detail images
   */
  images?: ModelImage[];
  status?: ModelGoodsStatus;
  stock?: number;
  title?: string;
}

export enum ModelGoodsStatus {
  GoodsOnSale = 1,
  GoodsOffSale = 2,
  GoodsOutOfStock = 3,
}

export type ModelImage = {
  src?: string;
  title?: string;
}

export type ModelList<T> = T[]

export type ModelListGoodsResponse = {
  items?: ModelGoodsInfo[];
  /*
   * @description Url of next page. If there is no more items, nextPage field not exists.
   */
  nextPage?: string;
}

export type ModelMap<T extends string | number, V> = Record<T, V>

export type ModelMultipleParamGeneric<T, V> = {
  A?: T;
  B?: V;
}

export type ModelSampleGenericType<T> = {
  array?: ModelArray<T>;
  list?: ModelList<T>;
  map?: ModelMap<number, T>;
  multipleParamGeneric?: ModelMultipleParamGeneric<number, T>;
  selfRef?: ModelGenericTypeResponse<T>;
  value?: T;
}

export type ModelUpdateGoodsRequest = {
  /*
   * @description Url of cover image
   */
  cover?: string;
  /*
   * @description Detail images
   */
  images?: ModelImage[];
  status?: ModelGoodsStatus;
  stock?: number;
  title?: string;
}

export type ModelUploadFileRes = {
  id?: number;
  url?: string;
}
