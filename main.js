// 多语言字典
const i18n = {
  zh: {
    navFeatures: "核心特性",
    navQuickStart: "快速开始",
    navGithub: "GitHub 源码",
    badgeText: "v1.0.0 稳定版正式发布 &bull; 开源自托管",
    heroTitle: "轻松部署与运营你的 <span class='text-gradient'>Xray 正向代理</span> 服务",
    heroSubtitle: "Volans 是专为极简运维打造的开源可视化代理控制台。只需点几下鼠标，即可把一台普通服务器变成安全、稳定、可随时分发订阅的节点集群。",
    btnQuickStart: "快速开始",
    btnGithub: "查看 GitHub",
    cmdLabel: "一键快速安装命令",
    copySuccess: "已复制到剪贴板！",
    featuresTag: "CAPABILITIES",
    featuresTitle: "为现代代理运维而生的全部特性",
    featuresDesc: "抛弃繁琐难懂的 JSON 配置文件，享受开箱即用、安全透明的全新管理体验。",
    feat1Title: "可视化内核管理",
    feat1Desc: "图形化管理 Xray 核心运行状态，零门槛配置 VLESS 协议、监听端口、TCP 及 WebSocket 传输流。",
    feat2Title: "订阅用户与配额管理",
    feat2Desc: "支持创建独立订阅客户，精细分配总流量配额、上下行独立限速，以及设置自定义到期时间。",
    feat3Title: "实时流量监控",
    feat3Desc: "内置 WebSocket 双向低延迟通信，即时推送每秒上/下行网络吞吐速率及历史流量曲线。",
    feat4Title: "通用订阅一键导出",
    feat4Desc: "自动生成标准订阅 Token 与客户端连接，一键复制直连，主流客户端无缝即刻接入。",
    feat5Title: "安全默认设计",
    feat5Desc: "内置 TLS 传输加密与域名嗅探 (Sniffing) 防护，严格的 JWT 身份鉴权与密码哈希保护。",
    feat6Title: "现代极简控制台",
    feat6Desc: "对标 Vercel / Grafana 设计哲学，支持无缝亮暗色切换，低饱和度科技美学，极致专注操作。",
    quickStartTag: "GETTING STARTED",
    quickStartTitle: "三步完成部署上线",
    quickStartDesc: "支持主流 Linux 发行版，单二进制独立运行，无需复杂的第三方运行环境。",
    step1Num: "01. 下载二进制",
    step1Title: "一键下载发行版",
    step1Desc: "从 GitHub Releases 获取对应架构的预编译二进制执行文件并赋予运行权限：",
    step2Num: "02. 配置守护进程",
    step2Title: "Systemd 后台托管",
    step2Desc: "配置为系统级服务，实现开机自启、崩溃自动重启与标准化日志捕获：",
    step3Num: "03. 访问控制台",
    step3Title: "打开浏览器初始化",
    step3Desc: "在浏览器中输入服务器公网 IP，登录默认管理员账号并开启您的第一个节点：",
    langToggle: "English"
  },
  en: {
    navFeatures: "Features",
    navQuickStart: "Quick Start",
    navGithub: "GitHub",
    badgeText: "v1.0.0 Stable Released &bull; Open Source & Self-hosted",
    heroTitle: "Effortlessly Deploy & Manage Your <span class='text-gradient'>Xray Proxy Node</span>",
    heroSubtitle: "Volans is an open-source visual management dashboard tailored for lightweight proxy operations. Transform any server into a secure, distributable subscription node with a few clicks.",
    btnQuickStart: "Get Started",
    btnGithub: "View on GitHub",
    cmdLabel: "Quick Install Command",
    copySuccess: "Copied to clipboard!",
    featuresTag: "CAPABILITIES",
    featuresTitle: "Built for Modern Proxy Operations",
    featuresDesc: "Say goodbye to error-prone JSON configuration files and enjoy a frictionless, transparent control panel.",
    feat1Title: "Visual Core Management",
    feat1Desc: "Monitor Xray core status in real time. Configure VLESS protocols, listening ports, TCP, and WebSocket transports seamlessly.",
    feat2Title: "Subscription & User Quotas",
    feat2Desc: "Create independent subscriber accounts with bandwidth quotas, independent uplink/downlink speed limits, and expiration dates.",
    feat3Title: "Real-time Traffic Telemetry",
    feat3Desc: "Powered by native WebSockets for ultra-low latency push of live throughput metrics and historical traffic graphs.",
    feat4Title: "Instant Subscription Export",
    feat4Desc: "Generate standard subscription tokens and client URLs with one-click copy support for universal client import.",
    feat5Title: "Secure by Default",
    feat5Desc: "Native TLS encryption and domain sniffing protection with hardened JWT authentication and password hashing.",
    feat6Title: "Modern Minimal UI",
    feat6Desc: "Inspired by Vercel & Grafana console aesthetics, with seamless Dark/Light mode switching and focused design.",
    quickStartTag: "GETTING STARTED",
    quickStartTitle: "Deploy in 3 Simple Steps",
    quickStartDesc: "Runs as a standalone binary on modern Linux distributions without heavyweight runtime dependencies.",
    step1Num: "01. Download Binary",
    step1Title: "Fetch Latest Release",
    step1Desc: "Download the prebuilt standalone binary from GitHub Releases and grant execution permissions:",
    step2Num: "02. Systemd Service",
    step2Title: "Manage as a Daemon",
    step2Desc: "Configure systemd for auto-start on boot, zero-downtime recovery, and standardized logging:",
    step3Num: "03. Access Web UI",
    step3Title: "Initialize Dashboard",
    step3Desc: "Open your browser to the server IP on port 8080 and log in with default credentials:",
    langToggle: "中文"
  }
};

let currentLang = 'zh';

// 切换语言
function toggleLanguage() {
  currentLang = currentLang === 'zh' ? 'en' : 'zh';
  updateLanguageUI();
}

function updateLanguageUI() {
  const dict = i18n[currentLang];
  document.querySelectorAll('[data-i18n]').forEach((el) => {
    const key = el.getAttribute('data-i18n');
    if (dict[key]) {
      el.innerHTML = dict[key];
    }
  });
}

// 复制文本到剪贴板
async function copyCommand(text) {
  try {
    if (navigator.clipboard && window.isSecureContext) {
      await navigator.clipboard.writeText(text);
    } else {
      const textArea = document.createElement('textarea');
      textArea.value = text;
      textArea.style.position = 'fixed';
      textArea.style.opacity = '0';
      document.body.appendChild(textArea);
      textArea.focus();
      textArea.select();
      document.execCommand('copy');
      document.body.removeChild(textArea);
    }
    const btn = document.getElementById('copy-cmd-btn');
    const originalText = btn.innerHTML;
    btn.innerHTML = `<svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 6L9 17l-5-5"/></svg> ${i18n[currentLang].copySuccess}`;
    setTimeout(() => {
      btn.innerHTML = originalText;
    }, 2000);
  } catch (err) {
    console.error('Copy failed', err);
  }
}

// 星空粒子背景绘制
function initStarfield() {
  const canvas = document.getElementById('canvas-starfield');
  if (!canvas) return;
  const ctx = canvas.getContext('2d');
  let stars = [];
  const numStars = 75;

  function resize() {
    canvas.width = window.innerWidth;
    canvas.height = window.innerHeight;
    stars = [];
    for (let i = 0; i < numStars; i++) {
      stars.push({
        x: Math.random() * canvas.width,
        y: Math.random() * canvas.height,
        radius: Math.random() * 1.2 + 0.3,
        alpha: Math.random() * 0.7 + 0.2,
        speed: Math.random() * 0.005 + 0.002
      });
    }
  }

  function render() {
    ctx.clearRect(0, 0, canvas.width, canvas.height);
    for (let i = 0; i < stars.length; i++) {
      const star = stars[i];
      star.alpha += star.speed;
      if (star.alpha > 0.9 || star.alpha < 0.2) {
        star.speed = -star.speed;
      }
      ctx.beginPath();
      ctx.arc(star.x, star.y, star.radius, 0, Math.PI * 2);
      ctx.fillStyle = `rgba(165, 180, 252, ${star.alpha})`;
      ctx.fill();
    }
    requestAnimationFrame(render);
  }

  window.addEventListener('resize', resize);
  resize();
  render();
}

window.addEventListener('DOMContentLoaded', () => {
  initStarfield();
  updateLanguageUI();
});
