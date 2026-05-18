-- 测试用户（密码 test1234 的 bcrypt）
INSERT INTO users (id, email, password_hash, created_at)
VALUES ('33333333-3333-3333-3333-333333333333', 'sft-tester@example.com',
        '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', NOW());

-- 4 场面试
INSERT INTO interviews (id, user_id, started_at, ended_at, status)
VALUES ('44444444-4444-4444-4444-444444444444', '33333333-3333-3333-3333-333333333333',
        NOW() - INTERVAL '4 hours', NOW() - INTERVAL '3 hour 30 min', 'finished'),
       ('55555555-5555-5555-5555-555555555555', '33333333-3333-3333-3333-333333333333',
        NOW() - INTERVAL '3 hours', NOW() - INTERVAL '2 hour 30 min', 'finished'),
       ('66666666-6666-6666-6666-666666666666', '33333333-3333-3333-3333-333333333333',
        NOW() - INTERVAL '2 hours', NOW() - INTERVAL '1 hour 30 min', 'finished'),
       ('77777777-7777-7777-7777-777777777777', '33333333-3333-3333-3333-333333333333',
        NOW() - INTERVAL '1 hour', NOW() - INTERVAL '30 min', 'finished');

-- 20 条 interview_turns
INSERT INTO interview_turns (id, interview_id, turn_id, stage, question, user_answer, asr_raw, created_at)
VALUES
    -- Go 后端面试
    (gen_random_uuid(), '44444444-4444-4444-4444-444444444444', 'T01', 'intro',
     '请简单介绍一下你自己以及最近的 Go 项目经验。',
     '我叫张三，三年 Go 后端经验，最近在做高并发订单系统，主力栈 Hertz + Kitex。',
     '我叫 张三 三年 Go 后端经验 最近在做高并发订单系统', NOW() - INTERVAL '3 hour 55 min'),
    (gen_random_uuid(), '44444444-4444-4444-4444-444444444444', 'T02', 'questioning',
     'GMP 调度模型里 P 的核心职责是什么？', 'P 是逻辑处理器，持有可运行 G 队列；M 必须绑定 P 才能执行 G。',
     'P 是逻辑处理器 持有可运行 G 队列', NOW() - INTERVAL '3 hour 50 min'),
    (gen_random_uuid(), '44444444-4444-4444-4444-444444444444', 'T03', 'questioning',
     '一个未关闭的 channel 在 select 里会怎样表现？', '未关闭的 channel 在 select 里阻塞等待发送/接收，不会触发零值返回。',
     '未关闭的 channel 在 select 里阻塞等待', NOW() - INTERVAL '3 hour 45 min'),
    (gen_random_uuid(), '44444444-4444-4444-4444-444444444444', 'T04', 'questioning', '怎么排查线上 goroutine 泄漏？',
     '抓 pprof goroutine profile，看 goroutine 数量趋势和阻塞栈。', 'pprof goroutine profile 看 数量趋势',
     NOW() - INTERVAL '3 hour 40 min'),
    (gen_random_uuid(), '44444444-4444-4444-4444-444444444444', 'T05', 'closing', '你还有什么想了解我们团队的？',
     '想了解一下技术栈演进规划，以及代码评审的协作方式。', '技术栈演进规划 代码评审', NOW() - INTERVAL '3 hour 35 min'),

    -- Java 后端面试
    (gen_random_uuid(), '55555555-5555-5555-5555-555555555555', 'T01', 'intro',
     '请介绍一下你的 Java 项目，重点说一个有挑战的部分。',
     '我做过支付清算系统，最有挑战的是分布式事务，用 Seata AT 模式 + 对账兜底。', '支付清算 分布式事务 Seata AT',
     NOW() - INTERVAL '2 hour 55 min'),
    (gen_random_uuid(), '55555555-5555-5555-5555-555555555555', 'T02', 'questioning',
     'HashMap 在并发场景下会出现什么问题？', '1.7 头插会成环导致死循环，1.8 改尾插但仍可能丢数据。',
     'HashMap 1.7 头插 死循环', NOW() - INTERVAL '2 hour 50 min'),
    (gen_random_uuid(), '55555555-5555-5555-5555-555555555555', 'T03', 'questioning', 'Spring 事务在哪些情况下会失效？',
     '同类自调用、私有方法、异常被吞、传播级别错配、checked exception 等。', '同类自调用 私有方法 异常被吞',
     NOW() - INTERVAL '2 hour 45 min'),
    (gen_random_uuid(), '55555555-5555-5555-5555-555555555555', 'T04', 'questioning', 'CMS 和 G1 的核心区别？',
     'CMS 标记清除碎片严重；G1 区域化划分可预测 STW，8u40+ 默认推荐。', 'CMS 标记清除 G1 区域化 STW',
     NOW() - INTERVAL '2 hour 40 min'),
    (gen_random_uuid(), '55555555-5555-5555-5555-555555555555', 'T05', 'closing',
     '你对加入我们后第一年的目标有什么期待？', '先把现有清算系统稳定性提升，再参与下一代架构设计。',
     '稳定性提升 下一代架构', NOW() - INTERVAL '2 hour 35 min'),

    -- 前端面试
    (gen_random_uuid(), '66666666-6666-6666-6666-666666666666', 'T01', 'intro', '请介绍一下你做过的前端项目。',
     '做过一个 React 数据看板，二期接入了 WebSocket 实时推送。', 'React 数据看板 WebSocket',
     NOW() - INTERVAL '1 hour 55 min'),
    (gen_random_uuid(), '66666666-6666-6666-6666-666666666666', 'T02', 'questioning',
     'React 的 useEffect 在 strict mode 下为什么会执行两次？',
     '是 React 18 开发期主动行为，检测不安全副作用，生产构建只执行一次。', 'React 18 strict mode 副作用',
     NOW() - INTERVAL '1 hour 50 min'),
    (gen_random_uuid(), '66666666-6666-6666-6666-666666666666', 'T03', 'questioning',
     '浏览器事件循环里 microtask 和 macrotask 的顺序？', '一次 macrotask 结束后清空所有 microtask 才会渲染并进下一轮。',
     'macrotask microtask 渲染', NOW() - INTERVAL '1 hour 45 min'),
    (gen_random_uuid(), '66666666-6666-6666-6666-666666666666', 'T04', 'questioning', '怎么优化首屏加载时间？',
     '路由懒加载 + 关键 CSS 内联 + 图片 webp + service worker 缓存。', '路由懒加载 CSS webp SW',
     NOW() - INTERVAL '1 hour 40 min'),
    (gen_random_uuid(), '66666666-6666-6666-6666-666666666666', 'T05', 'closing', '团队前端的技术栈和你熟悉的差异大吗？',
     '看了你们用 React 19 + Vite，差不多；Server Component 我还在学。', 'React 19 Vite Server Component',
     NOW() - INTERVAL '1 hour 35 min'),

    -- 算法面试
    (gen_random_uuid(), '77777777-7777-7777-7777-777777777777', 'T01', 'intro', '简单介绍你对算法工程化方向的兴趣。',
     '本科 ACM 区域赛铜牌，研究生方向召回排序，希望算法落地工程化。', 'ACM 召回排序 工程化', NOW() - INTERVAL '55 min'),
    (gen_random_uuid(), '77777777-7777-7777-7777-777777777777', 'T02', 'questioning',
     '快速排序最坏情况是什么？怎么缓解？', 'O(n²) 出现在已排序输入；三数取中或随机化 pivot 可降低概率。',
     '快排 O(n²) 三数取中', NOW() - INTERVAL '50 min'),
    (gen_random_uuid(), '77777777-7777-7777-7777-777777777777', 'T03', 'questioning', '动态规划和分治的核心区别？',
     '分治子问题独立；DP 子问题有重叠，要记忆化或自底向上填表。', 'DP 重叠 记忆化', NOW() - INTERVAL '45 min'),
    (gen_random_uuid(), '77777777-7777-7777-7777-777777777777', 'T04', 'questioning', '如何在亿级别数据中找 Top 100？',
     '小顶堆 + 流式扫描，空间 O(K)，时间 O(N log K)；分布式分片归并。', 'Top K 小顶堆 归并', NOW() - INTERVAL '40 min'),
    (gen_random_uuid(), '77777777-7777-7777-7777-777777777777', 'T05', 'closing', '反问环节，你想问什么？',
     '团队对模型部署的依赖深不深？是用 K8s 还是有更轻的方案？', '模型部署 K8s', NOW() - INTERVAL '35 min');

-- 16 条 questionnaire_results（每场 4 条，T05 故意不标）
INSERT INTO questionnaire_results
    (id, interview_id, turn_id, quality, feedback, created_at)
VALUES (gen_random_uuid(), '44444444-4444-4444-4444-444444444444', 'T01', 'good', '开场提问自然，给候选人热身机会。',
        NOW()),
       (gen_random_uuid(), '44444444-4444-4444-4444-444444444444', 'T02', 'good',
        '考察基础到位，追问 M-P 绑定逻辑可以更深一层。', NOW()),
       (gen_random_uuid(), '44444444-4444-4444-4444-444444444444', 'T03', 'bad', '提问太书本，缺少对实际死锁场景的延伸。',
        NOW()),
       (gen_random_uuid(), '44444444-4444-4444-4444-444444444444', 'T04', 'good', '从答案直接引出 pprof 实战，方向对路。',
        NOW()),
       (gen_random_uuid(), '55555555-5555-5555-5555-555555555555', 'T01', 'good', '让候选人挑挑战点说，了解项目深度。',
        NOW()),
       (gen_random_uuid(), '55555555-5555-5555-5555-555555555555', 'T02', 'bad',
        '只问 HashMap 太基础，应该结合 CHM 分段锁演进。', NOW()),
       (gen_random_uuid(), '55555555-5555-5555-5555-555555555555', 'T03', 'good', '事务失效场景覆盖完整。', NOW()),
       (gen_random_uuid(), '55555555-5555-5555-5555-555555555555', 'T04', 'bad',
        'CMS 和 G1 偏八股，可加场景化判断「何时选 ZGC」。', NOW()),
       (gen_random_uuid(), '66666666-6666-6666-6666-666666666666', 'T01', 'good', '让候选人主动展开项目细节，节奏好。',
        NOW()),
       (gen_random_uuid(), '66666666-6666-6666-6666-666666666666', 'T02', 'good',
        'strict mode 是好切入点，能筛掉死记硬背。', NOW()),
       (gen_random_uuid(), '66666666-6666-6666-6666-666666666666', 'T03', 'bad',
        '事件循环答得通顺但没追问浏览器和 Node 的差异。', NOW()),
       (gen_random_uuid(), '66666666-6666-6666-6666-666666666666', 'T04', 'good',
        '首屏优化结合 HTTP/2 和 service worker 视野广。', NOW()),
       (gen_random_uuid(), '77777777-7777-7777-7777-777777777777', 'T01', 'good', '从背景引出兴趣方向，问得很温和。',
        NOW()),
       (gen_random_uuid(), '77777777-7777-7777-7777-777777777777', 'T02', 'bad',
        '快排最坏问得过于经典，背得出来不代表理解。', NOW()),
       (gen_random_uuid(), '77777777-7777-7777-7777-777777777777', 'T03', 'good', 'DP vs 分治这个角度合适。', NOW()),
       (gen_random_uuid(), '77777777-7777-7777-7777-777777777777', 'T04', 'bad',
        'Top K 应该追问外部排序而不是停留小顶堆。', NOW());