import { request } from "umi";
import { 
  ViewGoodsCreateReq,
  ViewGoodsCreateRes,
  ViewGoodsDownRes,
  ViewGoodsInfoRes
 } from "./types";

/*
 * @description 创建商品接口
 */
export function shopGoodsCreate(data: ViewGoodsCreateReq) {
  return request<ViewGoodsCreateRes>(`/api/goods`, {
    method: "post",
    data,
  });
}

/*
 * @description 删除商品
 */
export function shopGoodsDelete(guid: string, query: { formDataField?: string }) {
  return request(`/api/goods/${guid}`, {
    method: "delete",
    params: query,
  });
}

/*
 * @description 下架商品
 */
export function shopGoodsDown(guid: string, query: { defaultQuery?: string }, data: { dateRange?: string[]; defaultPostForm?: string; operatorUid?: string; }) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => formData.append(key, data[key]));
  return request<ViewGoodsDownRes>(`/api/goods/${guid}/down`, {
    method: "post",
    params: query,
    data: formData,
  });
}

/*
 * @description 商品详情
 */
export function shopGoodsInfo(guid: string) {
  return request<ViewGoodsInfoRes>(`/api/v2/goods/${guid}`, {
    method: "get",
  });
}

export function shopWrappedHandler(query: { hello?: string; world?: string }) {
  return request<{ code: number; data: Record<string, any>; msg: string; }>(`/wrapped-handler`, {
    method: "get",
    params: query,
  });
}