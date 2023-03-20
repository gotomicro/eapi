import { request } from "umi";
import {
  ViewGoodsCreateReq,
  CustomResponseType,
  ViewGoodsCreateRes,
  ShopGoodsDownRequest,
  ViewGoodsDownRes,
  ViewGoodsInfoRes
} from "./types";

/*
 * @description GoodsCreate 创建商品接口
 */
export function shopGoodsCreate(data: ViewGoodsCreateReq) {
  return request<CustomResponseType<ViewGoodsCreateRes>>(`/api/goods`, {
    method: "post",
    data,
  });
}

/*
 * @description GoodsDelete 删除商品
 */
export function shopGoodsDelete(guid: string, query: { formDataField?: string }) {
  return request(`/api/goods/${guid}`, {
    method: "delete",
    params: query,
  });
}

/*
 * @description GoodsDown 下架商品
 */
export function shopGoodsDown(guid: string, query: { defaultQuery?: string }, data: ShopGoodsDownRequest) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => formData.append(key, data[key]));
  return request<ViewGoodsDownRes>(`/api/goods/${guid}/down`, {
    method: "post",
    params: query,
    data: formData,
  });
}

/*
 * @description GoodsInfo 商品详情
 */
export function shopGoodsInfo(guid: string) {
  return request<ViewGoodsInfoRes>(`/api/v2/goods/${guid}`, {
    method: "get",
  });
}

/*
 * @description wrapped handler
 */
export function shopWrappedHandler(query: { hello?: string; world?: string }) {
  return request<CustomResponseType<Record<string, any>>>(`/wrapped-handler`, {
    method: "get",
    params: query,
  });
}