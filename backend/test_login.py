import requests

s = requests.Session()
res = s.post('http://localhost:5000/api/v1/auth/login', json={"username":"afari", "password":"password"})
print("Login:", res.status_code, res.text)
if res.status_code == 200:
    res2 = s.get('http://localhost:5000/api/v1/auth/session')
    print("Session:", res2.status_code, res2.text)
