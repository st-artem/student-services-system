import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
  vus: 5,
  duration: '30s',
  thresholds: {
    http_req_failed: ['rate<0.01'], // Помилок менше 1%
    http_req_duration: ['p(95)<200'], // 95% запитів швидші за 200ms
  },
};

export default function () {
  // Звертаємося через port-forward (див. команди запуску)
  const res = http.get('http://localhost:8080/api/v1/items');
  check(res, {
    'status is 200': (r) => r.status === 200,
  });
  sleep(1);
}