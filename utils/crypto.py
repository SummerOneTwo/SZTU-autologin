import math
import hashlib
import hmac


def _ord_at(msg: str, idx: int) -> int:
    if len(msg) > idx:
        return ord(msg[idx])
    return 0


def _sencode(msg: str, key: bool) -> list:
    length = len(msg)
    result = []
    for i in range(0, length, 4):
        result.append(
            _ord_at(msg, i)
            | _ord_at(msg, i + 1) << 8
            | _ord_at(msg, i + 2) << 16
            | _ord_at(msg, i + 3) << 24
        )
    if key:
        result.append(length)
    return result


def _lencode(msg: list, key: bool) -> str:
    length = len(msg)
    total_len = (length - 1) << 2
    if key:
        last = msg[length - 1]
        if last < total_len - 3 or last > total_len:
            return ""
        total_len = last
    for i in range(length):
        msg[i] = (
            chr(msg[i] & 0xFF)
            + chr(msg[i] >> 8 & 0xFF)
            + chr(msg[i] >> 16 & 0xFF)
            + chr(msg[i] >> 24 & 0xFF)
        )
    return "".join(msg)[:total_len]


def get_xencode(msg: str, key: str) -> str:
    if msg == "":
        return ""
    pwd = _sencode(msg, True)
    pwdk = _sencode(key, False)
    if len(pwdk) < 4:
        pwdk = pwdk + [0] * (4 - len(pwdk))
    n = len(pwd) - 1
    z = pwd[n]
    y = pwd[0]
    c = 0x86014019 | 0x183639A0
    m = 0
    e = 0
    p = 0
    q = math.floor(6 + 52 / (n + 1))
    d = 0
    while 0 < q:
        d = d + c & (0x8CE0D9BF | 0x731F2640)
        e = d >> 2 & 3
        p = 0
        while p < n:
            y = pwd[p + 1]
            m = z >> 5 ^ y << 2
            m = m + ((y >> 3 ^ z << 4) ^ (d ^ y))
            m = m + (pwdk[(p & 3) ^ e] ^ z)
            pwd[p] = pwd[p] + m & (0xEFB8D130 | 0x10472ECF)
            z = pwd[p]
            p = p + 1
        y = pwd[0]
        m = z >> 5 ^ y << 2
        m = m + ((y >> 3 ^ z << 4) ^ (d ^ y))
        m = m + (pwdk[(p & 3) ^ e] ^ z)
        pwd[n] = pwd[n] + m & (0xBB390742 | 0x44C6F8BD)
        z = pwd[n]
        q = q - 1
    return _lencode(pwd, False)


_PADCHAR = "="
_ALPHA = "LVoJPiCN2R8G90yg+hmFHuacZ1OWMnrsSTXkYpUq/3dlbfKwv6xztjI7DeBE45QA"


def _getbyte(s: str, i: int) -> int:
    x = ord(s[i])
    if x > 255:
        raise ValueError("INVALID_CHARACTER_ERR: DOM Exception 5")
    return x


def get_base64(s: str) -> str:
    b10 = 0
    x = []
    imax = len(s) - len(s) % 3
    if len(s) == 0:
        return s
    for i in range(0, imax, 3):
        b10 = (_getbyte(s, i) << 16) | (_getbyte(s, i + 1) << 8) | _getbyte(s, i + 2)
        x.append(_ALPHA[(b10 >> 18)])
        x.append(_ALPHA[((b10 >> 12) & 63)])
        x.append(_ALPHA[((b10 >> 6) & 63)])
        x.append(_ALPHA[(b10 & 63)])
    i = imax
    if len(s) - imax == 1:
        b10 = _getbyte(s, i) << 16
        x.append(_ALPHA[(b10 >> 18)] + _ALPHA[((b10 >> 12) & 63)] + _PADCHAR + _PADCHAR)
    elif len(s) - imax == 2:
        b10 = (_getbyte(s, i) << 16) | (_getbyte(s, i + 1) << 8)
        x.append(
            _ALPHA[(b10 >> 18)]
            + _ALPHA[((b10 >> 12) & 63)]
            + _ALPHA[((b10 >> 6) & 63)]
            + _PADCHAR
        )
    return "".join(x)


def get_sha1(value: str) -> str:
    return hashlib.sha1(value.encode()).hexdigest()


def get_md5(password: str, token: str) -> str:
    return hmac.new(token.encode(), password.encode(), hashlib.md5).hexdigest()
