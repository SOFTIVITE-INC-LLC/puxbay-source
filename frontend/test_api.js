const http = require('http');

const loginData = JSON.stringify({
  email: "admin@thinkce.com",
  password: "password123"
});

const req = http.request('http://localhost:8080/api/v1/auth/login', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
    'Content-Length': loginData.length
  }
}, (res) => {
  let body = '';
  res.on('data', chunk => body += chunk);
  res.on('end', () => {
    const data = JSON.parse(body);
    const token = data.access_token || data.token;
    
    if (!token) {
        console.log("NO TOKEN", body);
        return;
    }

    http.get('http://localhost:8080/api/v1/products', {
      headers: { 'Authorization': 'Bearer ' + token }
    }, (res2) => {
      let body2 = '';
      res2.on('data', chunk => body2 += chunk);
      res2.on('end', () => {
        console.log(JSON.stringify(JSON.parse(body2).data[0], null, 2));
      });
    });
  });
});

req.write(loginData);
req.end();
