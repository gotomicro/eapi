import axios, { AxiosRequestConfig } from "axios";
import {
  ModelListGoodsResponse,
  ModelUpdateGoodsRequest,
  ModelGoodsInfo,
  ModelCreateGoodsRequest,
  ModelGenericTypeResponse,
  UploaderUploadFileRequest,
  ModelUploadFileRes
} from "./types";

/*
 * @description List
 */
export function goodsList(query: { since?: string; limit: number }, config?: AxiosRequestConfig) {
  return axios.get<ModelListGoodsResponse>(`/v1/goods`, {
    params: query,
    ...config,
  });
}

/*
 * @description Update
 */
export function goodsUpdate(data: ModelUpdateGoodsRequest, config?: AxiosRequestConfig) {
  return axios.patch<ModelGoodsInfo[]>(`/v1/goods`, {
    data,
    ...config,
  });
}

/*
 * @description Create
 */
export function goodsCreate(data: ModelCreateGoodsRequest, config?: AxiosRequestConfig) {
  return axios.post<ModelGoodsInfo>(`/v1/goods`, {
    data,
    ...config,
  });
}

/*
 * @description Delete
 */
export function goodsDelete(id: string, config?: AxiosRequestConfig) {
  return axios.delete<ModelGenericTypeResponse<string>>(`/v1/goods/${id}`, {
    ...config,
  });
}

/*
 * @description UploadFile
 */
export function uploaderUploadFile(data: UploaderUploadFileRequest, config?: AxiosRequestConfig) {
  const formData = new FormData();
  Object.keys(data).forEach((key) => formData.append(key, data[key]));
  return axios.post<ModelUploadFileRes>(`/v1/upload`, {
    data: formData,
    ...config,
  });
}