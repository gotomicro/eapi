import { request } from "umi";
import { 
  ViewGoodsCreateReq,
  GinParams,
  ViewSelfRefType,
  ViewProperty,
  ViewGoodsInfoRes
 } from "./types";

/*
 * @description 创建商品接口
 */
export function shopGoodsCreate(data: ViewGoodsCreateReq) {
  return request<{ Status?: GinParams; guid?: string; raw?: any; selfRef?: ViewSelfRefType; stringAlias?: string; }>(`/api/goods`, {
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
export function shopGoodsDown(guid: string, data: { dateRange?: string[]; operatorUid?: string; }) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => formData.append(key, data[key]));
  return request<{ Status?: string; }>(`/api/goods/${guid}/down`, {
    method: "post",
    data: formData,
  });
}

/*
 * @description 商品详情
 */
export function shopGoodsInfo(guid: string) {
  return request<{ cover?: string; mapInt?: Record<number, ViewProperty>; price?: number; properties?: Record<string, ViewProperty>; subTitle?: string; title?: string; }>(`/api/v2/goods/${guid}`, {
    method: "get",
  });
}

export function shopWrappedHandler(query: { hello?: string; world?: string }) {
  return request<{ code: number; data: ViewGoodsInfoRes; msg: string; }>(`/wrapped-handler`, {
    method: "get",
    params: query,
  });
}