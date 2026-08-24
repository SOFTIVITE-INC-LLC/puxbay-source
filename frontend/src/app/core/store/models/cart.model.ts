import { Product } from './product.model';

export interface CartItem {
  product_id: string;
  quantity: number;
  product?: Product; // populated locally for details
}

export interface CartResponse {
  cart: CartItem[];
}

export interface CartPayload {
  product_id: string;
  quantity: number;
}
