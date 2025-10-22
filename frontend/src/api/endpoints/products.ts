// src/api/endpoints/products.ts

import { getApiEndpoint, registerEndpoints, ApiEndpoint, EndpointConfig } from './_core';

const productsEndpoints: { [action: string]: EndpointConfig } = {
  "list": { "url": "/products", "method": "GET" }
};

registerEndpoints('products', productsEndpoints);

// --- 导出的函数 ---

export function listProductsEndpoint(): ApiEndpoint<'GET'> {
  return getApiEndpoint('products', 'list');
}