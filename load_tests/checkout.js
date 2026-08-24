import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 10, // Virtual Users
    duration: '30s',
};

export default function () {
    const url = 'http://localhost:8080/api/v1/storefront/checkout';
    const payload = JSON.stringify({
        cart_items: [
            { product_id: '123e4567-e89b-12d3-a456-426614174000', quantity: 2 },
        ],
        customer_email: 'loadtest@example.com',
    });

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'X-Tenant': 'demo',
        },
    };

    const res = http.post(url, payload, params);
    check(res, {
        'is status 200 or 400 (validation)': (r) => r.status === 200 || r.status === 400,
    });
    sleep(1);
}
