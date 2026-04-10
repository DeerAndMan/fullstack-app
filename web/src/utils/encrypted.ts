// RSA 加密
import { pki, util, random } from "node-forge";

export const publicKeyPem = `-----BEGIN CERTIFICATE-----
MIIEojCCA4qgAwIBAgIUJnDfZpkqvqVe2IgcuIVLepADM/MwDQYJKoZIhvcNAQEL
BQAwgYsxCzAJBgNVBAYTAlVTMRkwFwYDVQQKExBDbG91ZEZsYXJlLCBJbmMuMTQw
MgYDVQQLEytDbG91ZEZsYXJlIE9yaWdpbiBTU0wgQ2VydGlmaWNhdGUgQXV0aG9y
aXR5MRYwFAYDVQQHEw1TYW4gRnJhbmNpc2NvMRMwEQYDVQQIEwpDYWxpZm9ybmlh
MB4XDTI1MDIyNjAyMjkwMFoXDTQwMDIyMzAyMjkwMFowYjEZMBcGA1UEChMQQ2xv
dWRGbGFyZSwgSW5jLjEdMBsGA1UECxMUQ2xvdWRGbGFyZSBPcmlnaW4gQ0ExJjAk
BgNVBAMTHUNsb3VkRmxhcmUgT3JpZ2luIENlcnRpZmljYXRlMIIBIjANBgkqhkiG
9w0BAQEFAAOCAQ8AMIIBCgKCAQEAtnx47eYKWWT/41nmb/FLocFU36cx+k7GUt4d
nYkkme9PlzxlsQMvZiIfVKR7o9Wp9Hd4Mh4jormF7gJ45oLqDkr6OpImqLwe5U1H
75fNuFGzDzo47wzV7vrYrh4Uqri5HpG974HI2GA8tVrNLxQSa2CionseNxK2yJ0t
Z2LkZ8gwg2Q2LtqxrM83A+W8WwddjvlWasiFDonYaFALGSonQ8sYapfBIEIDWRzl
samJD6/ZBKHQqBZ1d7fmT1DZp87kLLCr8mv7Fq3mmDo3d2YlrEAEfH1xh1GBpDHh
tLvTl9UW7BN2HP9ksPs6bPtXk5OWi7IaCbNeYJrQsSqQlh+AnwIDAQABo4IBJDCC
ASAwDgYDVR0PAQH/BAQDAgWgMB0GA1UdJQQWMBQGCCsGAQUFBwMCBggrBgEFBQcD
ATAMBgNVHRMBAf8EAjAAMB0GA1UdDgQWBBSF0e2rss2hi2f4QDEiW9hFjjlxLTAf
BgNVHSMEGDAWgBQk6FNXXXw0QIep65TbuuEWePwppDBABggrBgEFBQcBAQQ0MDIw
MAYIKwYBBQUHMAGGJGh0dHA6Ly9vY3NwLmNsb3VkZmxhcmUuY29tL29yaWdpbl9j
YTAlBgNVHREEHjAcgg0qLjk5OTkxMTAueHl6ggs5OTk5MTEwLnh5ejA4BgNVHR8E
MTAvMC2gK6AphidodHRwOi8vY3JsLmNsb3VkZmxhcmUuY29tL29yaWdpbl9jYS5j
cmwwDQYJKoZIhvcNAQELBQADggEBAL/sxzBX2V97lFSuLTsfnje8gK/LADbEuqmZ
HOE+cf1MpInYKXqcQ6dnDK3ND7NjIpVp6CX5a8DKoYnL7cTtoesptsmuzwT78xwE
0wmA7wAvpYQuNtxS7lJSONEi6TVoV98X1yTOJZihF+LimO7Yl4n7DKs+wRrztziy
1M0or9eTcerCxSd7q3/E/sBPoPilMqkQBR0o3MoHk11sy++dcmXqnNLEwNHKOlOv
v9HCsSzoPUnrAK7eyI71t/KtoY6Anne4f35sZ7O+iBVNBkyaDWJLZeZafuENAEP9
an0+I2Y1dTt2YTW/5jTpha8yNYwVBSi9ahA32SWzrwOyTVjA28U=
-----END CERTIFICATE-----`;

/**
 * 使用公钥加密密码
 * @param password 密码
 * @returns base64Password 加密后的密码
 */
export const encryptPassword = (password: string, saltLen?: number): string => {
  const saltStr = generateSalt(saltLen);
  // 从证书中提取公钥
  const certificate = pki.certificateFromPem(publicKeyPem);
  // 获取公钥
  const publicKey = certificate.publicKey as pki.rsa.PublicKey;
  // 使用公钥加密密码
  const encrypted = publicKey.encrypt(password + saltStr, "RSA-OAEP");
  const base64Password = util.encode64(encrypted);

  // console.log(
  //   "原始加密！！！",
  //   util.encode64(publicKey.encrypt(password, "RSA-OAEP"))
  // );
  // console.log("base64Password", base64Password);

  return base64Password;
};

/**
 * 生成盐
 * @param length 长度
 * @returns salt 盐
 */
export const generateSalt = (length: number = 16): string => {
  const salt = random.getBytesSync(length);
  return util.encode64(salt);
};
