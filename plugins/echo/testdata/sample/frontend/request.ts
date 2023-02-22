import { request } from "umi";
import { 
  ModelListGoodsResponse,
  ModelCreateGoodsRequest,
  ModelGoodsInfo,
  ModelUpdateGoodsRequest,
  ModelGenericTypeResponse,
  UploaderUploadFileRequest,
  ModelUploadFileRes
 } from "./types";

/*
 * @description List
 */
export function goodsList(query: { since?: string; limit?: number }) {
  return request<ModelListGoodsResponse>(`/v1/goods`, {
    method: "get",
    params: query,
  });
}

/*
 * @description Create
 */
export function goodsCreate(data: ModelCreateGoodsRequest) {
  return request<ModelGoodsInfo>(`/v1/goods`, {
    method: "post",
    data,
  });
}

/*
 * @description Update
 */
export function goodsUpdate(data: ModelUpdateGoodsRequest) {
  return request<ModelGoodsInfo[]>(`/v1/goods`, {
    method: "patch",
    data,
  });
}

/*
 * @description Delete
 */
export function goodsDelete(id: string) {
  return request<ModelGenericTypeResponse<string>>(`/v1/goods/${id}`, {
    method: "delete",
  });
}

/*
 * @description UploadFile
 */
export function uploaderUploadFile(data: UploaderUploadFileRequest) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => formData.append(key, data[key]));
  return request<ModelUploadFileRes>(`/v1/upload`, {
    method: "post",
    data: formData,
  });
}