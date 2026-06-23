/**
 * 超星学习通 - 自动完成视频任务点
 * 用法：在 Edge 课程页面按 F12 → Console → 粘贴整段代码 → 回车
 * 停止：在控制台输入 chaoxingAuto.stop()
 */
(function () {
  'use strict';

  if (window.chaoxingAuto) {
    console.log('[超星助手] 已在运行，重新启动...');
    window.chaoxingAuto.stop();
  }

  const CONFIG = {
    playbackRate: 1.5,       // 倍速（未完成任务点禁止快进，倍速一般可用）
    checkInterval: 1000,     // 视频状态检查间隔(ms)
    minWatchRatio: 0.91,     // 观看进度阈值（任务点要求≥90%）
    scrollStep: 500,         // 每节内下滑步长(px)
    loadWait: 3500,          // 切换章节后等待加载(ms)
    betweenTaskWait: 2000,   // 任务点之间等待(ms)
    betweenSectionWait: 2500 // 切换小节后等待(ms)
  };

  const state = {
    running: true,
    timer: null,
    currentVideo: null,
    handledVideos: new WeakSet()
  };

  function log(msg, color = '#4CAF50') {
    console.log(`%c[超星助手] ${msg}`, `color:${color};font-weight:bold`);
  }

  function warn(msg) {
    log(msg, '#FF9800');
  }

  function err(msg) {
    log(msg, '#F44336');
  }

  function sleep(ms) {
    return new Promise(r => setTimeout(r, ms));
  }

  // ---- iframe / 视频查找 ----

  function getContentFrame() {
    try {
      if (window.frames['iframe']) return window.frames['iframe'];
      const f = document.querySelector('iframe#iframe, iframe[name="iframe"], iframe.ans-insert-content-online');
      return f ? f.contentWindow : null;
    } catch (e) {
      return null;
    }
  }

  function getContentDoc() {
    const win = getContentFrame();
    return win ? win.document : null;
  }

  function findVideosInDoc(doc) {
    const videos = [];
    if (!doc) return videos;
    try {
      doc.querySelectorAll('video').forEach(v => videos.push(v));
      doc.querySelectorAll('iframe').forEach(f => {
        try {
          const inner = f.contentDocument || f.contentWindow?.document;
          if (inner) inner.querySelectorAll('video').forEach(v => videos.push(v));
        } catch (_) {}
      });
    } catch (_) {}
    return videos;
  }

  function findAllVideos() {
    const found = [];
    const seen = new Set();
    const add = (v) => {
      if (v && !seen.has(v)) {
        seen.add(v);
        found.push(v);
      }
    };

    document.querySelectorAll('video').forEach(add);
    document.querySelectorAll('iframe').forEach(f => {
      try {
        findVideosInDoc(f.contentDocument || f.contentWindow?.document).forEach(add);
      } catch (_) {}
    });

    const contentDoc = getContentDoc();
    if (contentDoc) findVideosInDoc(contentDoc).forEach(add);

    return found;
  }

  function getUnfinishedTaskBlocks(doc) {
    if (!doc) return [];
    const blocks = [...doc.querySelectorAll('.ans-attach-ct')];
    return blocks.filter(b => {
      const finished = b.classList.contains('ans-job-finished') ||
        b.querySelector('.ans-job-finished') ||
        b.closest('.ans-job-finished');
      const hasJob = b.querySelector('.ans-job-icon, .orangeNew, [class*="job"]');
      return !finished && (hasJob || b.querySelector('iframe[src*="video"]'));
    });
  }

  function scrollContentDown() {
    const doc = getContentDoc();
    const targets = [
      doc?.documentElement,
      doc?.body,
      document.documentElement,
      document.body,
      document.querySelector('.scroll-content, .main-content, .content, #content, .chapter-content')
    ].filter(Boolean);

    for (const el of targets) {
      const before = el.scrollTop;
      el.scrollTop += CONFIG.scrollStep;
      if (el.scrollTop > before) return true;
      el.scrollBy?.(0, CONFIG.scrollStep);
      if (el.scrollTop > before) return true;
    }
    window.scrollBy(0, CONFIG.scrollStep);
    return true;
  }

  // ---- 视频播放 ----

  async function tryPlayVideo(video) {
    if (!video || state.handledVideos.has(video)) return false;

    const duration = video.duration;
    if (!duration || !isFinite(duration) || duration <= 0) {
      // 尝试点击播放按钮
      const doc = video.ownerDocument;
      const playBtn = doc.querySelector('.vjs-big-play-button, .vjs-play-control, button[title="播放"]');
      playBtn?.click();
      await sleep(1500);
      if (!video.duration || !isFinite(video.duration)) return false;
    }

    video.muted = true;
    try {
      video.playbackRate = CONFIG.playbackRate;
      await video.play();
      state.currentVideo = video;
      log(`开始播放视频 (${Math.round(video.duration)}秒, ${CONFIG.playbackRate}x)`);
      return true;
    } catch (e) {
      warn('播放失败，尝试静音重播...');
      video.muted = true;
      try {
        await video.play();
        state.currentVideo = video;
        return true;
      } catch (e2) {
        return false;
      }
    }
  }

  function waitVideoComplete(video) {
    return new Promise(resolve => {
      const start = Date.now();
      const maxMs = ((video.duration || 600) / CONFIG.playbackRate + 30) * 1000;

      const tick = () => {
        if (!state.running) return resolve('stopped');

        const dur = video.duration;
        const cur = video.currentTime;

        if (video.ended || (dur > 0 && cur / dur >= CONFIG.minWatchRatio)) {
          state.handledVideos.add(video);
          log(`视频完成 (${Math.round(cur)}/${Math.round(dur)}秒)`);
          return resolve('done');
        }

        if (video.paused && !video.ended) {
          video.play().catch(() => {});
        }

        if (Date.now() - start > maxMs) {
          warn('超时，视为完成');
          state.handledVideos.add(video);
          return resolve('timeout');
        }

        state.timer = setTimeout(tick, CONFIG.checkInterval);
      };

      video.addEventListener('ended', () => resolve('ended'), { once: true });
      tick();
    });
  }

  // ---- 当前节内：处理所有视频任务点 ----

  async function processCurrentSectionVideos() {
    log('扫描当前小节视频任务点...');
    let rounds = 0;
    const maxRounds = 30;

    while (state.running && rounds < maxRounds) {
      rounds++;
      const contentDoc = getContentDoc();
      const taskBlocks = getUnfinishedTaskBlocks(contentDoc);
      const allVideos = findAllVideos();
      const pendingVideos = allVideos.filter(v => !state.handledVideos.has(v));

      if (pendingVideos.length === 0 && taskBlocks.length === 0) {
        // 再下滑看看有没有隐藏的视频
        scrollContentDown();
        await sleep(800);
        const moreVideos = findAllVideos().filter(v => !state.handledVideos.has(v));
        if (moreVideos.length === 0) {
          log('当前小节视频任务已全部处理');
          return true;
        }
      }

      let played = false;
      for (const video of pendingVideos) {
        video.scrollIntoView?.({ behavior: 'smooth', block: 'center' });
        await sleep(500);
        if (await tryPlayVideo(video)) {
          await waitVideoComplete(video);
          played = true;
          await sleep(CONFIG.betweenTaskWait);
          scrollContentDown();
          await sleep(800);
          break;
        }
      }

      if (!played) {
        if (taskBlocks.length > 0) {
          taskBlocks[0].scrollIntoView?.({ behavior: 'smooth', block: 'center' });
          scrollContentDown();
          await sleep(1000);
        } else {
          scrollContentDown();
          await sleep(1000);
        }
      }
    }

    warn('达到最大扫描轮次，进入下一节');
    return true;
  }

  // ---- 侧边栏：查找并点击下一未完成小节 ----

  function getSidebarItems() {
    const selectors = [
      '#coursetree .posCatalog_select:not(.firstLayer) .posCatalog_name',
      '#coursetree .posCatalog_select:not(.firstLayer)',
      '[class*="catalog"] [class*="item"]',
      '[class*="Catalog"] [class*="name"]',
      '.catalog_points .catalog_title',
      '.catalog_points li',
      '.chapter-item',
      '.node-item'
    ];

    for (const sel of selectors) {
      const items = [...document.querySelectorAll(sel)];
      if (items.length > 0) return items;
    }
    return [];
  }

  function isItemIncomplete(el) {
    const node = el.closest('li, [class*="catalog"], [class*="Catalog"], [class*="item"]') || el;
    const text = node.textContent || '';

    // 绿色勾 = 已完成
    if (node.querySelector('.icon_Completed, .icon-finish, [class*="finish"], [class*="complete"]')) {
      return false;
    }
    if (node.classList.contains('icon_Completed')) return false;

    // 橙色数字徽章 = 有未完成任务点
    const badge = node.querySelector('[class*="orange"], [class*="badge"], [class*="tip"], .jobCount, .tips');
    if (badge) {
      const n = parseInt(badge.textContent.trim(), 10);
      if (n === 0) return false;
      if (n > 0) return true;
    }

    // 当前激活项且页面有未完成任务
    if (node.classList.contains('posCatalog_active') || node.closest('.posCatalog_active')) {
      return findAllVideos().some(v => !state.handledVideos.has(v)) ||
        getUnfinishedTaskBlocks(getContentDoc()).length > 0;
    }

    // 无完成标记的小节视为待处理
    const parent = node.closest('li') || node;
    if (parent.querySelector('.icon_Completed, .icon-finish')) return false;

    return true;
  }

  function getCurrentSidebarIndex(items) {
    return items.findIndex(el => {
      const node = el.closest('.posCatalog_active, [class*="active"], [class*="current"]') || el;
      return node.classList.contains('posCatalog_active') ||
        node.closest('.posCatalog_active') ||
        el.classList.contains('active') ||
        el.closest('[class*="active"]');
    });
  }

  function clickSidebarItem(el) {
    const clickable = el.closest('.posCatalog_select')?.querySelector('.posCatalog_name') || el;
    const title = clickable.getAttribute?.('title') || clickable.textContent?.trim() || '未知';
    log(`切换到: ${title}`, '#2196F3');
    clickable.click();
    return title;
  }

  async function goNextSection() {
    // 方式1：经典下一节按钮
    const nextBtn = document.querySelector('#prevNextFocusNext, .nextChapter, [class*="next"]');
    if (nextBtn && nextBtn.offsetParent !== null) {
      log('点击"下一节"按钮');
      nextBtn.click();
      await sleep(CONFIG.loadWait);
      state.handledVideos = new WeakSet();
      return true;
    }

    // 方式2：侧边栏目录
    const items = getSidebarItems();
    if (items.length === 0) {
      err('找不到侧边栏目录，请手动点击下一节后重新运行脚本');
      return false;
    }

    const curIdx = getCurrentSidebarIndex(items);
    log(`目录共 ${items.length} 项，当前索引: ${curIdx}`);

    // 从当前位置往后找第一个未完成项
    for (let i = Math.max(0, curIdx + 1); i < items.length; i++) {
      if (isItemIncomplete(items[i])) {
        items[i].scrollIntoView?.({ behavior: 'smooth', block: 'center' });
        await sleep(500);
        clickSidebarItem(items[i]);
        await sleep(CONFIG.loadWait);
        state.handledVideos = new WeakSet();
        return true;
      }
    }

    // 从头再找（可能当前节后面的都完成了）
    for (let i = 0; i < items.length; i++) {
      if (i === curIdx) continue;
      if (isItemIncomplete(items[i])) {
        items[i].scrollIntoView?.({ behavior: 'smooth', block: 'center' });
        await sleep(500);
        clickSidebarItem(items[i]);
        await sleep(CONFIG.loadWait);
        state.handledVideos = new WeakSet();
        return true;
      }
    }

    log('🎉 恭喜！所有章节任务点已处理完毕！', '#4CAF50');
    return false;
  }

  // ---- 防暂停 ----

  function installAntiPause() {
    const resume = () => {
      if (state.currentVideo?.paused) {
        state.currentVideo.play().catch(() => {});
      }
    };
    document.addEventListener('visibilitychange', resume);
    window.addEventListener('blur', resume);
    return () => {
      document.removeEventListener('visibilitychange', resume);
      window.removeEventListener('blur', resume);
    };
  }

  // ---- 主循环 ----

  async function mainLoop() {
    const cleanup = installAntiPause();
    log('=== 自动刷课启动 ===', '#4CAF50');
    log(`倍速: ${CONFIG.playbackRate}x | 进度阈值: ${CONFIG.minWatchRatio * 100}%`);
    log('停止命令: chaoxingAuto.stop()');

    while (state.running) {
      try {
        await processCurrentSectionVideos();
        if (!state.running) break;

        await sleep(CONFIG.betweenSectionWait);
        const hasNext = await goNextSection();
        if (!hasNext) break;

      } catch (e) {
        err(`出错: ${e.message}，3秒后重试...`);
        await sleep(3000);
      }
    }

    cleanup();
    log('=== 自动刷课已停止 ===', '#9E9E9E');
  }

  window.chaoxingAuto = {
    stop() {
      state.running = false;
      if (state.timer) clearTimeout(state.timer);
      log('正在停止...', '#9E9E9E');
    },
    restart() {
      state.running = true;
      state.handledVideos = new WeakSet();
      mainLoop();
    },
    config: CONFIG
  };

  mainLoop();
})();
