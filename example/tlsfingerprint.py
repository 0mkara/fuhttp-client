import socket
import sys
import json
import random
from retrying import retry
import base64
from utils import getCookieStr, aquireSocket
import uuid

with open('user_agents.json') as u:
    agents = json.load(u)

def retry_on_timeout(exc):
    print("will retry if Timeout")
    print(type(exc))
    return isinstance(exc, ValueError)

def getTLSFingerprint(proxy, agent: str, parrotId: int):
    sock = aquireSocket()
    session = str(uuid.uuid4())
    # Send data
    headers = {
                "Host": "client.tlsfingerprint.io",
                "Connection": "keep-alive",
                "User-Agent": agent
            }
    options = {
                "url": "https://client.tlsfingerprint.io:8443",
                "proxy": proxy,
                "gzip": "true",
                "headers": headers,
                "header_order": 'GET / HTTP/1.1\r\n' + '\r\n'.join("{!s}: {!s}".format(key,val) for (key,val) in headers.items()) + '\r\n\r\n',
                "parrotId": parrotId,
                "session_id": session
            }
    message = json.dumps(options).encode('utf-8')
    print('sending {!r}'.format(message))
    sock.sendall(message)

    data = sock.recv(4*4096)
    print('received {!r}'.format(data))
    sock.close()
    res = json.loads(data)
    if res['error']:
        raise ValueError(res['error'])
    if res['response']['statusCode'] != 200:
        raise ValueError(res['response']['statusCode'])
    body = base64.b64decode(res['body'])
    return body.decode("utf-8")

class Dsg():
    agent = agents['Chrome_std_linux'][5]
    parrotId = 10

    def run(self):
        try:
            fp = getTLSFingerprint(None, self.agent, self.parrotId)
            print("Fingerprint: ")
            print(fp)
        except Exception as e:
            print(e)
            sys.exit(0)
        finally:
            print('Done')
            sys.exit(0)
if __name__ == "__main__":
    Dsg().run()