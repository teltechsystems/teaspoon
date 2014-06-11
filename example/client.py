import teaspoon

tsp = teaspoon.Teaspoon('127.0.0.1', 8000)
response = tsp.send_request(teaspoon.Request(method=0x0, resource=0x0, payload='HELLO'))

print response.payload