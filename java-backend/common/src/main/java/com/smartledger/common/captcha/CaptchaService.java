package com.smartledger.common.captcha;

import org.springframework.stereotype.Service;

import java.awt.Color;
import java.awt.Font;
import java.awt.Graphics2D;
import java.awt.RenderingHints;
import java.awt.image.BufferedImage;
import java.io.ByteArrayOutputStream;
import java.security.SecureRandom;
import java.util.Base64;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import javax.imageio.ImageIO;

/**
 * 图形验证码：内存存储，单节点够用。
 * <p>
 * 生产多副本部署时应改为 Redis 等集中存储。
 */
@Service
public class CaptchaService {

    private static final int TTL_MS = 5 * 60 * 1000;
    private static final SecureRandom RANDOM = new SecureRandom();

    private final Map<String, Entry> store = new ConcurrentHashMap<>();

    public Captcha generate() {
        String code = randomCode(4);
        String id = Long.toHexString(RANDOM.nextLong()) + Long.toHexString(RANDOM.nextLong());
        store.put(id, new Entry(code.toLowerCase(), System.currentTimeMillis() + TTL_MS));
        String image = renderBase64(code);
        return new Captcha(id, image);
    }

    /**
     * @param clear 校验成功后是否立即删除，登录场景传 true 防止重放
     */
    public boolean verify(String id, String answer, boolean clear) {
        if (id == null || answer == null) {
            return false;
        }
        Entry entry = store.get(id);
        if (entry == null || entry.expiresAt < System.currentTimeMillis()) {
            store.remove(id);
            return false;
        }
        boolean ok = entry.code.equals(answer.trim().toLowerCase());
        if (ok && clear) {
            store.remove(id);
        }
        return ok;
    }

    /** 去掉易混淆字符（0/O、1/l 等），降低误输率 */
    private static String randomCode(int len) {
        String chars = "23456789abcdefghjkmnpqrstuvwxyz";
        StringBuilder sb = new StringBuilder(len);
        for (int i = 0; i < len; i++) {
            sb.append(chars.charAt(RANDOM.nextInt(chars.length())));
        }
        return sb.toString();
    }

    /** 返回裸 Base64，由前端自行加 data URI 前缀 */
    private static String renderBase64(String code) {
        int w = 120;
        int h = 40;
        BufferedImage img = new BufferedImage(w, h, BufferedImage.TYPE_INT_RGB);
        Graphics2D g = img.createGraphics();
        g.setColor(new Color(240, 244, 248));
        g.fillRect(0, 0, w, h);
        g.setRenderingHint(RenderingHints.KEY_ANTIALIASING, RenderingHints.VALUE_ANTIALIAS_ON);
        g.setFont(new Font(Font.SANS_SERIF, Font.BOLD, 22));
        g.setColor(new Color(30, 41, 59));
        g.drawString(code, 18, 28);
        for (int i = 0; i < 6; i++) {
            g.setColor(new Color(RANDOM.nextInt(180), RANDOM.nextInt(180), RANDOM.nextInt(180)));
            g.drawLine(RANDOM.nextInt(w), RANDOM.nextInt(h), RANDOM.nextInt(w), RANDOM.nextInt(h));
        }
        g.dispose();
        try {
            ByteArrayOutputStream baos = new ByteArrayOutputStream();
            ImageIO.write(img, "png", baos);
            return Base64.getEncoder().encodeToString(baos.toByteArray());
        } catch (Exception e) {
            throw new IllegalStateException("captcha render failed", e);
        }
    }

    private record Entry(String code, long expiresAt) {}

    public record Captcha(String captchaId, String image) {}
}
