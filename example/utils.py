from http import cookies
import socket
import sys
from datetime import datetime, timedelta
from urllib.parse import quote
import uuid


def getCookies(data):
    try:
        ck = cookies.SimpleCookie()
        for c in data:
            ck.load(c)
        return ck
    except Exception as e:
        raise e


def getCookieStr(ck):
    try:
        # create cookies string
        cks = "; ".join([str(x)+"="+str(y.value) for x, y in ck.items()])
        return cks
    except Exception as e:
        raise e


def appendCookie(cookies, cookie):
    try:
        cookies.load(cookie)
        return cookies
    except Exception as e:
        raise e

def appendCookies(cookies, newCookies):
    try:
        for c in newCookies:
            cookies.load(c)
        return cookies
    except Exception as e:
        raise e

def aquireSocket():
    # Create a UDS socket
    sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)
    # Connect the socket to the port where the server is listening
    server_address = '/tmp/fuhttp.sock'
    print('connecting to {}'.format(server_address))
    try:
        sock.connect(server_address)
        return sock
    except socket.error as msg:
        print(msg)
        sys.exit(1)