import { describe, expect, it } from "vitest";
import { pki, util } from "node-forge";

import { encryptPassword, generateSalt, publicKeyPem } from "@/utils/encrypted";

describe("encrypted utils", () => {
  it("按指定字节长度生成 Base64 盐值", () => {
    const salt = generateSalt(24);

    expect(util.decode64(salt)).toHaveLength(24);
  });

  it("默认生成 16 字节盐值", () => {
    expect(util.decode64(generateSalt())).toHaveLength(16);
  });

  it("使用证书公钥生成可解码的 RSA 密文", () => {
    const encrypted = encryptPassword("P@ssw0rd", 8);
    const certificate = pki.certificateFromPem(publicKeyPem);
    const publicKey = certificate.publicKey as pki.rsa.PublicKey;

    expect(encrypted).toMatch(/^[A-Za-z0-9+/]+={0,2}$/);
    expect(util.decode64(encrypted)).toHaveLength(Math.ceil(publicKey.n.bitLength() / 8));
  });

  it("相同密码每次加密结果不同", () => {
    expect(encryptPassword("same-password", 8)).not.toBe(encryptPassword("same-password", 8));
  });
});
