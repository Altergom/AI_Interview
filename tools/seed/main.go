// tools/seed/main.go 题库种子数据导入脚本。
// 用法：go run tools/seed/main.go
// 会向 PG 写入 bank_questions 并投递 vectorize_task 消息到 RabbitMQ。
// 依赖：PG / RabbitMQ 已启动（docker compose up -d）
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/joho/godotenv"

	"ai_interview/internal/config"
	"ai_interview/internal/domain"
	"ai_interview/internal/log"
	"ai_interview/internal/mq"
	"ai_interview/internal/mq/mqclient"
	"ai_interview/internal/storage/postgres"
)

func main() {
	_ = godotenv.Load(".env", ".env.local")
	if err := config.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	// 连接 PG
	db, err := postgres.New(ctx, postgres.Options{
		DSN:             config.Cfg.PostgresDSN,
		MaxOpenConns:    config.Cfg.PGMaxOpenConns,
		MaxIdleConns:    config.Cfg.PGMaxIdleConns,
		ConnMaxLifetime: config.Cfg.PGConnMaxLifetime,
		ConnMaxIdleTime: config.Cfg.PGConnMaxIdleTime,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect pg: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	repo := postgres.NewBankQuestionRepo(db.Gorm())

	// 连接 MQ
	mqCli, err := mqclient.New(config.Cfg.MQBrokerURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "connect mq: %v\n", err)
		os.Exit(1)
	}
	defer mqCli.Close()

	if err := mqCli.DeclareQueue(mq.TopicVectorizeTask); err != nil {
		fmt.Fprintf(os.Stderr, "declare queue: %v\n", err)
		os.Exit(1)
	}

	questions := seedQuestions()
	log.Infof("[Seed] importing %d questions", len(questions))

	ok, fail := 0, 0
	for _, q := range questions {
		id, err := repo.Insert(ctx, q)
		if err != nil {
			log.Errorf("[Seed] insert failed question=%q: %v", q.Question[:min(30, len(q.Question))], err)
			fail++
			continue
		}
		task := mq.VectorizeTask{QuestionID: id}
		if err := mqCli.Publish(ctx, mq.TopicVectorizeTask, task); err != nil {
			log.Errorf("[Seed] publish vectorize_task id=%s: %v", id, err)
		}
		ok++
	}
	log.Infof("[Seed] done: inserted=%d failed=%d", ok, fail)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// seedQuestions 返回覆盖 5 个方向的种子题目（约 80 道）。
func seedQuestions() []*domain.BankQuestionRecord {
	var questions []*domain.BankQuestionRecord

	// ---------- Go Backend ----------
	goBackend := []struct {
		q, a   string
		tags   []string
		diff   domain.Difficulty
	}{
		{
			q:    "Go 的 goroutine 和线程有什么区别？",
			a:    "goroutine 是用户态轻量线程，由 Go runtime 调度（M:N 模型），初始栈 2KB 可动态增长，创建/切换开销远低于 OS 线程；OS 线程栈固定（通常 MB 级），由内核调度。",
			tags: []string{"golang", "goroutine", "concurrency"},
			diff: domain.DifficultyEasy,
		},
		{
			q:    "channel 的底层实现原理是什么？",
			a:    "channel 底层是 hchan 结构体，包含环形缓冲区（buf）、发送/接收游标、等待队列（sendq/recvq）和互斥锁（lock）。发送时若 buf 满则 gopark 阻塞，接收方就绪后由 runtime 唤醒。",
			tags: []string{"golang", "channel", "concurrency"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Go 的 GC 算法是什么？如何调优？",
			a:    "Go 使用三色标记清除 + 混合写屏障，并发标记（无需 STW 扫描全部对象）。调优手段：GOGC 调大减少 GC 频率、GOMEMLIMIT 设置软内存上限、减少堆分配（复用对象、对象池）。",
			tags: []string{"golang", "gc", "performance"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "sync.Mutex 和 sync.RWMutex 有什么区别，何时使用 RWMutex？",
			a:    "Mutex 任何时刻只允许一个 goroutine 持有；RWMutex 允许多个并发读但写互斥。读多写少的场景（如缓存读取）用 RWMutex 可提高吞吐量；写频繁时 RWMutex 的读锁升级开销反而不划算。",
			tags: []string{"golang", "sync", "concurrency"},
			diff: domain.DifficultyEasy,
		},
		{
			q:    "Go interface 的底层结构是什么？nil interface 与包含 nil 指针的 interface 有什么区别？",
			a:    "interface 由 (itab *Itab, data unsafe.Pointer) 两个字段组成。nil interface 两个字段都为 nil；包含 nil 指针的 interface 的 itab 非 nil（有类型信息），因此 != nil，这是常见陷阱。",
			tags: []string{"golang", "interface", "internals"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "context 的作用是什么？如何正确传递和取消？",
			a:    "context 用于跨 API 边界传递截止时间、取消信号和请求范围的键值对。规范：作为第一个参数传入、不放进结构体、链式派生（WithCancel/WithTimeout/WithValue）、父 cancel 后所有子 context 自动取消。",
			tags: []string{"golang", "context", "cancellation"},
			diff: domain.DifficultyEasy,
		},
		{
			q:    "map 在并发场景下为什么不安全？有哪些解决方案？",
			a:    "Go 内置 map 并发读写会 panic（有竞态检测）。解决方案：1) sync.Mutex 包裹；2) sync.Map（适合读多写少、key 不重叠场景）；3) 分片锁；4) channel 串行化。",
			tags: []string{"golang", "map", "concurrency"},
			diff: domain.DifficultyEasy,
		},
		{
			q:    "Go 的 defer 执行顺序和性能影响？",
			a:    "defer 以 LIFO 顺序执行，在函数返回前触发（包括 panic 时）。Go 1.14 后 open-coded defer 优化使得无循环体内的 defer 几乎零开销；循环内的 defer 仍有堆分配代价，高频路径应避免。",
			tags: []string{"golang", "defer", "performance"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "如何实现一个高性能的 HTTP 服务，避免 goroutine 泄漏？",
			a:    "要点：1) 限制并发（worker pool 或信号量）；2) 所有阻塞调用接受 context 且正确处理 ctx.Done()；3) 读写超时配置（ReadTimeout/WriteTimeout）；4) goroutine 内 panic recover；5) 使用 pprof 定期检查 goroutine 数。",
			tags: []string{"golang", "http", "goroutine", "performance"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Go 泛型（generics）的使用场景和限制是什么？",
			a:    "泛型适用于：通用容器（集合、队列）、通用算法（排序、filter/map）、避免 interface{} 断言。限制：不支持方法上的类型参数、运行时反射不感知泛型类型、复杂约束可读性差，简单场景不应滥用泛型。",
			tags: []string{"golang", "generics"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "描述 Go 内存模型（happens-before）及其对编程的指导意义。",
			a:    "Go 内存模型规定：在 goroutine 内顺序一致；跨 goroutine 通过 channel 发送/接收、sync 原语建立 happens-before 关系。实际意义：对共享变量的读写必须通过同步原语保护，否则结果未定义（不保证可见性）。",
			tags: []string{"golang", "memory-model", "concurrency"},
			diff: domain.DifficultyHard,
		},
		{
			q:    "Go 中如何做 CPU/内存 profiling？",
			a:    "引入 net/http/pprof，访问 /debug/pprof；或使用 runtime/pprof 手动采样写文件。分析工具：go tool pprof + 火焰图（pprof -http 或 FlameGraph）。关注 allocs（内存分配）、goroutine（泄漏检测）、mutex（锁争用）。",
			tags: []string{"golang", "profiling", "performance"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "slice 和 array 的区别？append 的扩容机制是什么？",
			a:    "array 长度固定，值类型；slice 是 (ptr, len, cap) 的描述符，引用语义。append 在 cap 不足时扩容：Go 1.18+ 小 slice 2x 增长，大 slice 约 1.25x，底层分配新数组并 copy。扩容后原 slice 和新 slice 不共享底层数组。",
			tags: []string{"golang", "slice", "internals"},
			diff: domain.DifficultyEasy,
		},
		{
			q:    "Go 的 select 语句有哪些使用技巧？",
			a:    "1) 多 channel 同时就绪时随机选择（公平调度）；2) 配合 default 实现非阻塞收发；3) 配合 time.After/time.NewTimer 做超时；4) for-select 循环消费多路 channel；5) 空 select{} 永久阻塞（常用于 main goroutine 等待）。",
			tags: []string{"golang", "select", "channel"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Go 的 sync.Once 使用场景和原理？",
			a:    "sync.Once 保证函数只执行一次，适用于单例初始化。底层用原子操作 + 互斥锁：first-check 快路径（原子读）避免锁争用，已执行则直接返回；首次执行时加锁、执行函数、原子写 done=1。",
			tags: []string{"golang", "sync", "singleton"},
			diff: domain.DifficultyEasy,
		},
	}

	for _, q := range goBackend {
		questions = append(questions, &domain.BankQuestionRecord{
			Question:        q.q,
			StandardAnswer:  q.a,
			Tags:            q.tags,
			RelatedConcepts: []string{},
			FollowupQuestionIDs: []string{},
			Difficulty:      q.diff,
		})
	}

	// ---------- Java Backend ----------
	javaBackend := []struct {
		q, a   string
		tags   []string
		diff   domain.Difficulty
	}{
		{
			q:    "Java 中 HashMap 的底层实现原理是什么？Java 8 做了哪些优化？",
			a:    "HashMap 底层是数组+链表；Java 8 当链表长度 ≥ 8 且数组容量 ≥ 64 时链表转红黑树，查询退化为 O(log n)。扩容：容量 2x，重新 hash（高位运算优化，避免全量 rehash）。",
			tags: []string{"java", "hashmap", "collections"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "ConcurrentHashMap 和 Hashtable 有什么区别？",
			a:    "Hashtable 全表加 synchronized 锁，并发度低；ConcurrentHashMap（Java 8）用 CAS + synchronized 细粒度锁（仅锁 bucket 首节点），读操作无锁（volatile），并发度远高于 Hashtable。",
			tags: []string{"java", "concurrency", "hashmap"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "JVM 内存模型和垃圾回收器（G1/ZGC）的区别？",
			a:    "JVM 内存分堆（Eden/Survivor/Old）、方法区、栈、PC。G1 将堆分为大小相等的 Region，可预测停顿；ZGC 完全并发，STW < 1ms，适合低延迟场景；G1 适合吞吐量与延迟均衡场景。",
			tags: []string{"java", "jvm", "gc"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Java 线程池的核心参数和拒绝策略？",
			a:    "核心参数：corePoolSize、maximumPoolSize、keepAliveTime、workQueue、threadFactory。拒绝策略：AbortPolicy（抛异常，默认）、CallerRunsPolicy（调用者线程执行）、DiscardPolicy（静默丢弃）、DiscardOldestPolicy（丢弃队列最老任务）。",
			tags: []string{"java", "thread-pool", "concurrency"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "volatile 关键字的作用？能保证原子性吗？",
			a:    "volatile 保证可见性（写后立刻刷新到主内存、读时强制从主内存读）和有序性（禁止重排序），但不保证原子性（如 i++ 仍是三步操作）。复合操作需用 AtomicXxx 或 synchronized。",
			tags: []string{"java", "volatile", "concurrency"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Spring Bean 的生命周期？",
			a:    "实例化 → 属性注入 → BeanNameAware/BeanFactoryAware → BeanPostProcessor.before → @PostConstruct/InitializingBean.afterPropertiesSet → init-method → BeanPostProcessor.after → 使用 → @PreDestroy/DisposableBean.destroy → destroy-method。",
			tags: []string{"java", "spring", "bean-lifecycle"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Spring 的 AOP 实现原理？JDK 动态代理和 CGLIB 的区别？",
			a:    "Spring AOP 基于代理：目标类实现接口时默认用 JDK 动态代理（反射调用）；无接口时用 CGLIB 生成子类（字节码增强）。JDK 代理只能代理接口方法，CGLIB 可代理类方法但无法代理 final 方法。",
			tags: []string{"java", "spring", "aop", "proxy"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "MyBatis 的 #{} 和 ${} 有什么区别？",
			a:    "#{}  是预编译占位符，MyBatis 自动加引号并做 SQL 转义，防止 SQL 注入；${} 是字符串拼接，直接嵌入 SQL，存在注入风险，仅用于动态表名/列名等无法参数化的场景。",
			tags: []string{"java", "mybatis", "sql-injection"},
			diff: domain.DifficultyEasy,
		},
		{
			q:    "什么是双亲委派模型？为什么要打破它？",
			a:    "类加载时先委托父 ClassLoader 加载，确保核心类唯一性、防止篡改。打破场景：SPI（ContextClassLoader 反向委托）、OSGi（模块化隔离）、热部署（自定义 ClassLoader 实现类隔离重载）。",
			tags: []string{"java", "classloader", "jvm"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Java 中 synchronized 和 ReentrantLock 的区别？",
			a:    "synchronized 是 JVM 内置关键字，自动释放锁，不可中断等待；ReentrantLock 更灵活：可中断（lockInterruptibly）、可超时（tryLock）、公平锁、Condition 多条件变量。高竞争下 ReentrantLock 性能略优，简单场景 synchronized 更简洁。",
			tags: []string{"java", "lock", "concurrency"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "描述 Spring 事务传播机制中 REQUIRED 和 REQUIRES_NEW 的区别。",
			a:    "REQUIRED（默认）：有事务则加入，无则新建；REQUIRES_NEW：无论如何挂起外部事务并新建独立事务，内部事务提交/回滚不影响外部事务。适用场景：审计日志写入需独立提交，不受业务事务回滚影响时用 REQUIRES_NEW。",
			tags: []string{"java", "spring", "transaction"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Java 内存泄漏的常见原因和排查方法？",
			a:    "常见原因：静态集合持有对象引用、ThreadLocal 未 remove、事件监听器未注销、内部类持有外部类引用。排查：jmap -heap/jmap -histo 看堆分布；jstack 看线程；MAT（Eclipse Memory Analyzer）分析 heap dump；Arthas 在线诊断。",
			tags: []string{"java", "memory-leak", "jvm"},
			diff: domain.DifficultyHard,
		},
		{
			q:    "Java 8 Stream 的并行流（parallelStream）适合什么场景？有什么风险？",
			a:    "适合：大数据量、计算密集、无状态操作（filter/map）；不适合：I/O 密集（占用 ForkJoinPool）、有状态操作（collect 到非线程安全容器）、顺序依赖场景。风险：共享线程池影响其他并发任务，数据量小时开销反而更大。",
			tags: []string{"java", "stream", "parallel"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "什么是 CAS？ABA 问题如何解决？",
			a:    "CAS（Compare-And-Swap）是无锁原子操作：期望值匹配才更新，否则重试（自旋）。ABA 问题：值从 A→B→A，CAS 误以为未变更。解决：AtomicStampedReference 加版本号（stamp），每次修改 stamp 递增，比较时同时校验值和 stamp。",
			tags: []string{"java", "cas", "atomic"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "描述 Spring Cloud 中服务降级和熔断的区别（Resilience4j 为例）。",
			a:    "熔断（CircuitBreaker）：监控失败率，达阈值时 OPEN 状态直接快速失败，防止级联故障；半开后探测恢复。降级（Fallback）：调用失败/超时后执行兜底逻辑（返回默认值/缓存结果）。熔断决定是否调用，降级决定失败后怎么处理。",
			tags: []string{"java", "spring-cloud", "resilience"},
			diff: domain.DifficultyMedium,
		},
	}

	for _, q := range javaBackend {
		questions = append(questions, &domain.BankQuestionRecord{
			Question:            q.q,
			StandardAnswer:      q.a,
			Tags:                q.tags,
			RelatedConcepts:     []string{},
			FollowupQuestionIDs: []string{},
			Difficulty:          q.diff,
		})
	}

	// ---------- Frontend ----------
	frontend := []struct {
		q, a   string
		tags   []string
		diff   domain.Difficulty
	}{
		{
			q:    "React 的虚拟 DOM 和 Diff 算法是什么？",
			a:    "虚拟 DOM 是真实 DOM 的 JS 对象描述。React Diff（Reconciliation）基于三个假设：不同类型元素产生不同树；key 用于稳定标识元素；同层比较（O(n) 复杂度而非 O(n³)）。更新时只 patch 差异节点，减少重排/重绘。",
			tags: []string{"frontend", "react", "virtual-dom"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "useEffect 和 useLayoutEffect 的区别？",
			a:    "useEffect 在浏览器绘制后异步执行，不阻塞渲染，适合数据请求/订阅；useLayoutEffect 在 DOM 更新后、浏览器绘制前同步执行，适合需要读取/修改 DOM 布局（避免闪烁）的场景。服务端渲染时不要用 useLayoutEffect（无 DOM）。",
			tags: []string{"frontend", "react", "hooks"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "浏览器的事件循环（Event Loop）机制？",
			a:    "JS 单线程。执行栈清空后，先处理所有微任务队列（Promise.then/MutationObserver），再取一个宏任务（setTimeout/setInterval/I/O）执行，如此循环。微任务优先级高于宏任务，一轮宏任务后会清空全部微任务。",
			tags: []string{"frontend", "javascript", "event-loop"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "什么是闭包？有什么应用场景和潜在问题？",
			a:    "闭包是函数 + 其词法环境（外部变量引用）的组合。应用：防抖/节流、模块化封装、缓存（memoization）、事件处理绑定。潜在问题：循环中闭包捕获同一变量（用 let 或 IIFE 解决）、内存泄漏（闭包持有大对象引用）。",
			tags: []string{"frontend", "javascript", "closure"},
			diff: domain.DifficultyEasy,
		},
		{
			q:    "CSS 的 BFC（块格式化上下文）是什么？如何触发？",
			a:    "BFC 是独立的布局区域，内部元素不影响外部。触发条件：float 不为 none、overflow 不为 visible、display 为 inline-block/table-cell/flex/grid、position 为 absolute/fixed。作用：清除浮动、防止 margin 塌陷、防止文字环绕。",
			tags: []string{"frontend", "css", "bfc"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "前端性能优化的常见手段有哪些？",
			a:    "资源层：代码分割（lazy/Suspense）、Tree Shaking、压缩/Gzip/Brotli、CDN、图片 WebP/懒加载。渲染层：避免强制重排（批量 DOM 操作/requestAnimationFrame）、虚拟列表、useMemo/useCallback。网络层：HTTP/2 多路复用、预连接（preconnect）、Service Worker 缓存。",
			tags: []string{"frontend", "performance", "optimization"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "React 中 useMemo 和 useCallback 的区别及使用场景？",
			a:    "useMemo 缓存计算结果（值），适合昂贵计算；useCallback 缓存函数引用，适合传递给子组件的回调（配合 React.memo 避免子组件重渲染）。注意：过度使用增加内存压力，只在确有性能瓶颈时使用。",
			tags: []string{"frontend", "react", "hooks", "performance"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "HTTP 缓存机制（强缓存 vs 协商缓存）？",
			a:    "强缓存：Cache-Control: max-age / Expires，未过期直接读本地缓存（200 from cache），不请求服务器。协商缓存：过期后携带 If-None-Match（ETag）/ If-Modified-Since 请求服务器，未修改则 304 Not Modified，节省带宽。",
			tags: []string{"frontend", "http", "cache"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "XSS 和 CSRF 攻击原理及防御？",
			a:    "XSS：注入恶意脚本到页面执行。防御：输出转义（HTML Entity）、CSP、HttpOnly Cookie。CSRF：诱导用户向已登录站点发送伪造请求。防御：SameSite Cookie、CSRF Token（双重提交 Cookie 或随机 token 校验）、Referer 校验。",
			tags: []string{"frontend", "security", "xss", "csrf"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Webpack 的 Tree Shaking 原理？为什么只对 ES Module 有效？",
			a:    "Tree Shaking 依赖静态分析：ES Module 的 import/export 是静态声明（编译期可知依赖关系），打包时标记未使用的导出并在压缩阶段删除。CommonJS require() 是动态的（运行时才知道依赖），无法静态分析，所以不支持。",
			tags: []string{"frontend", "webpack", "tree-shaking"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "描述 Promise 的状态机及 async/await 的本质。",
			a:    "Promise 有三个状态：pending → fulfilled / rejected，状态不可逆。async/await 是 Promise 的语法糖：async 函数返回 Promise；await 暂停当前微任务，等待 Promise 解决后恢复，等价于 .then() 链式调用但代码更线性易读。",
			tags: []string{"frontend", "javascript", "promise", "async"},
			diff: domain.DifficultyEasy,
		},
		{
			q:    "React 18 的并发特性（Concurrent Mode）有哪些改进？",
			a:    "主要特性：Transitions（startTransition/useTransition）将非紧急更新标为可中断；Suspense 服务端流式渲染（Streaming SSR）；自动批处理（所有事件内的 setState 合并）；useId 生成稳定 SSR ID。核心：Fiber 架构支持中断/恢复渲染，保证交互响应优先。",
			tags: []string{"frontend", "react", "concurrent"},
			diff: domain.DifficultyHard,
		},
		{
			q:    "CSS-in-JS 和传统 CSS Modules 的取舍？",
			a:    "CSS-in-JS（styled-components/emotion）：动态样式方便、局部作用域、JS 逻辑与样式共存，缺点是运行时开销、SSR 复杂。CSS Modules：构建时静态化、性能好、零运行时，缺点是动态样式不便。大型项目追求性能选 CSS Modules 或 Tailwind；动态主题/组件库选 CSS-in-JS。",
			tags: []string{"frontend", "css", "styling"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "浏览器渲染流程（从 URL 到页面显示）？",
			a:    "DNS 解析 → TCP 握手 → TLS 握手 → HTTP 请求 → HTML 解析（DOM 树）→ CSSOM 构建 → Render Tree → Layout（计算位置/尺寸）→ Paint（绘制像素）→ Composite（合成层，GPU 加速）→ 显示。JS 执行可阻塞 HTML 解析（async/defer 优化）。",
			tags: []string{"frontend", "browser", "rendering"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "前端微服务（Micro Frontend）的实现方式有哪些？",
			a:    "主要方案：1) iframe（隔离最彻底但通信复杂）；2) Web Components（标准化，跨框架）；3) Module Federation（Webpack 5，共享依赖，动态加载子应用）；4) Single-SPA（路由劫持，多框架共存）。核心问题：样式隔离、JS 沙箱、路由协调、状态共享。",
			tags: []string{"frontend", "micro-frontend", "architecture"},
			diff: domain.DifficultyHard,
		},
	}

	for _, q := range frontend {
		questions = append(questions, &domain.BankQuestionRecord{
			Question:            q.q,
			StandardAnswer:      q.a,
			Tags:                q.tags,
			RelatedConcepts:     []string{},
			FollowupQuestionIDs: []string{},
			Difficulty:          q.diff,
		})
	}

	// ---------- Algorithm ----------
	algorithm := []struct {
		q, a   string
		tags   []string
		diff   domain.Difficulty
	}{
		{
			q:    "快速排序的原理和时间复杂度？如何避免最坏情况？",
			a:    "分治：选 pivot，分区为 [<pivot][pivot][>pivot]，递归排序子数组。平均 O(n log n)，最坏 O(n²)（有序数组+固定 pivot）。避免方式：随机选 pivot；三数取中（median-of-3）；小数组切换插入排序。",
			tags: []string{"algorithm", "sorting", "quicksort"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "动态规划和贪心算法的本质区别，如何判断用哪个？",
			a:    "DP：子问题存在重叠，需要状态转移记录最优子结构，全局最优依赖历史决策。贪心：每步选局部最优，需证明局部最优可导出全局最优（交换论证/拟阵）。判断方法：能证明贪心选择性质就用贪心（更高效）；否则 DP。",
			tags: []string{"algorithm", "dp", "greedy"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "BFS 和 DFS 各自的适用场景？",
			a:    "BFS：最短路径（无权图）、层序遍历、按距离扩展；用队列，空间 O(n)。DFS：全路径枚举、拓扑排序、连通分量、回溯剪枝；用栈/递归，空间 O(深度)。有环图 DFS 需标记访问状态；稠密图 BFS 内存压力大。",
			tags: []string{"algorithm", "bfs", "dfs", "graph"},
			diff: domain.DifficultyEasy,
		},
		{
			q:    "二叉树的各种遍历方式（前中后序、层序）及其应用场景？",
			a:    "前序（根左右）：序列化/复制树；中序（左根右）：BST 得有序序列；后序（左右根）：计算子树大小/删除节点；层序（BFS）：层级关系/最短路径。迭代实现：前中后序用显式栈模拟递归；层序用队列。",
			tags: []string{"algorithm", "tree", "traversal"},
			diff: domain.DifficultyEasy,
		},
		{
			q:    "描述红黑树的性质，为什么它能保证 O(log n)？",
			a:    "五条性质：节点红或黑；根黑；叶（NIL）黑；红节点的子节点黑（不连续红）；从任意节点到叶路径的黑节点数相同（黑高相等）。由黑高相等保证最短路径（全黑）不超过最长路径（红黑交替）的 2 倍，树高 ≤ 2log(n+1)，操作 O(log n)。",
			tags: []string{"algorithm", "red-black-tree", "data-structure"},
			diff: domain.DifficultyHard,
		},
		{
			q:    "什么是 LRU 缓存？如何实现 O(1) 的 get 和 put？",
			a:    "LRU（最近最少使用）：淘汰最久未被访问的 entry。O(1) 实现：HashMap 存 key→节点指针 + 双向链表维护访问顺序（最近用的移到头部，淘汰尾部）。HashMap 保证查找 O(1)，链表操作 O(1)（有指针直接操作）。",
			tags: []string{"algorithm", "lru", "cache", "data-structure"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "如何判断链表中是否有环？找到环的入口节点？",
			a:    "Floyd 判圈：快慢指针，快 2 步慢 1 步，相遇则有环。找入口：相遇后一个指针回头部，两个指针同步前进，再次相遇即入口（数学推导：设环外长 a，环内距入口 b，相遇点到入口 c，则 a = c）。",
			tags: []string{"algorithm", "linked-list", "two-pointers"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "最长公共子序列（LCS）和最长公共子串的区别及 DP 解法？",
			a:    "子序列不要求连续，子串要求连续。LCS：dp[i][j] = dp[i-1][j-1]+1（s[i]==s[j]）or max(dp[i-1][j], dp[i][j-1])，O(mn)。最长公共子串：dp[i][j] 只在匹配时累加，不匹配置 0，O(mn) 空间可优化为 O(min(m,n))。",
			tags: []string{"algorithm", "dp", "string"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Dijkstra 算法和 Bellman-Ford 的区别？各自的适用场景？",
			a:    "Dijkstra：贪心 + 优先队列，O((V+E)log V)，不支持负权边。Bellman-Ford：松弛 V-1 次，O(VE)，支持负权，可检测负环。选择：无负权用 Dijkstra；有负权用 Bellman-Ford 或 SPFA；需最短路到所有节点用 Floyd（O(V³)，可有负权无负环）。",
			tags: []string{"algorithm", "graph", "shortest-path"},
			diff: domain.DifficultyHard,
		},
		{
			q:    "什么是 B 树和 B+ 树？为什么数据库索引用 B+ 树？",
			a:    "B 树：多路平衡搜索树，内节点和叶节点都存数据。B+ 树：内节点只存键（更多键 → 更矮 → 更少 IO），数据全在叶节点，叶节点串联成链表支持范围查询。数据库用 B+ 树：1) 树更矮，IO 少；2) 叶链表支持高效范围扫描；3) 更稳定的磁盘读写模式。",
			tags: []string{"algorithm", "b-tree", "database", "index"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Top-K 问题的几种解法及复杂度对比？",
			a:    "1) 全排序 O(n log n)，取前 K；2) 堆（维护大小 K 的小根堆）O(n log K)，适合流式/在线场景；3) 快速选择（partition）平均 O(n)，最坏 O(n²)，不需排序；4) 计数排序 O(n+范围)，适合值域小的整数。工程选堆（稳定 O(n log K)，内存 O(K)）。",
			tags: []string{"algorithm", "top-k", "heap"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "如何设计一个支持 O(1) getRandom 的数据结构？",
			a:    "变长数组 + HashMap：数组存值，HashMap 存 value→index 映射。insert: 追加到数组尾，记录 index；remove: 将目标元素与尾元素交换，更新 HashMap，弹出尾部；getRandom: 随机下标访问数组。三个操作均摊 O(1)。",
			tags: []string{"algorithm", "data-structure", "design"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "字符串匹配的 KMP 算法原理？next 数组如何计算？",
			a:    "KMP 避免无效回溯：next[i] 是模式串 P[0..i] 的最长公共前后缀长度。匹配失败时，模式串不回到头部而是跳到 next[j-1] 位置继续。next 数组用双指针在 O(m) 内计算；整体复杂度 O(n+m)，优于朴素 O(nm)。",
			tags: []string{"algorithm", "string", "kmp"},
			diff: domain.DifficultyHard,
		},
		{
			q:    "归并排序和堆排序的对比，各自的优缺点？",
			a:    "归并排序：稳定，O(n log n) 保证，外排序（磁盘）首选，缺点是 O(n) 额外空间。堆排序：原地（O(1) 空间），O(n log n) 保证，不稳定，缓存不友好（访问模式跳跃，cache miss 高）。快排在实践中通常快于堆排，因缓存友好。",
			tags: []string{"algorithm", "sorting", "mergesort", "heapsort"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "并查集（Union-Find）的应用场景和路径压缩优化？",
			a:    "应用：连通分量判断、环检测、Kruskal 最小生成树。路径压缩：find 时把路径上所有节点直接挂到根（递归实现）。按秩合并：小树并入大树，树高 O(log n)。两个优化结合后 find 均摊接近 O(1)（逆阿克曼函数）。",
			tags: []string{"algorithm", "union-find", "data-structure"},
			diff: domain.DifficultyMedium,
		},
	}

	for _, q := range algorithm {
		questions = append(questions, &domain.BankQuestionRecord{
			Question:            q.q,
			StandardAnswer:      q.a,
			Tags:                q.tags,
			RelatedConcepts:     []string{},
			FollowupQuestionIDs: []string{},
			Difficulty:          q.diff,
		})
	}

	// ---------- AI Agent ----------
	aiAgent := []struct {
		q, a   string
		tags   []string
		diff   domain.Difficulty
	}{
		{
			q:    "什么是 RAG（Retrieval-Augmented Generation）？它解决了什么问题？",
			a:    "RAG 将检索与生成结合：先从外部知识库检索相关文档，再将其拼入 prompt 作为上下文，让 LLM 基于真实数据生成答案。解决了：1) LLM 知识截止问题；2) 幻觉（无依据捏造）；3) 私域知识无法注入；4) 减少微调成本。",
			tags: []string{"ai-agent", "rag", "llm"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "描述 ReAct（Reason + Act）框架的工作原理。",
			a:    "ReAct 让 LLM 交替输出 Thought（推理步骤）和 Action（工具调用），观察工具返回结果（Observation）后继续推理，形成循环直到给出 Final Answer。优势：推理链可审计、支持多步工具调用、比纯思维链更接地气。",
			tags: []string{"ai-agent", "react", "reasoning"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Function Calling 和 Tool Use 的本质是什么？如何保证 JSON 输出稳定？",
			a:    "本质：LLM 输出结构化的工具调用描述（函数名+参数 JSON），由框架层执行并将结果返回给 LLM。稳定性：1) 在 prompt 中提供清晰的 JSON schema；2) 设定 response_format=json_object；3) 加重试机制（解析失败重新要求生成）；4) 温度设低（0-0.2）减少随机性。",
			tags: []string{"ai-agent", "function-calling", "tool-use"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Agent 的 Memory 分类有哪些？各自适合什么场景？",
			a:    "1) In-context Memory：当前对话历史（短期，上下文窗口限制）；2) External Memory：向量数据库持久化（长期记忆，RAG 检索）；3) Entity Memory：提取关键实体存结构化存储（人物/事件关系）；4) Summary Memory：对话摘要压缩（平衡长期与窗口限制）。",
			tags: []string{"ai-agent", "memory", "llm"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "LLM 的 Prompt Engineering 有哪些核心技巧？",
			a:    "1) 角色设定（system prompt 明确身份）；2) 少样本（few-shot）示例；3) 思维链（Chain-of-Thought，step-by-step）；4) 结构化输出（JSON schema）；5) 任务分解（拆为子问题）；6) 负向指令（明确说不要做什么）；7) 温度控制（创造性 vs 确定性）。",
			tags: []string{"ai-agent", "prompt-engineering", "llm"},
			diff: domain.DifficultyEasy,
		},
		{
			q:    "向量数据库的选型考量（Milvus vs Qdrant vs Pinecone）？",
			a:    "Milvus：开源、支持多种索引（IVF/HNSW）、适合大规模（亿级）、k8s 友好，运维较重。Qdrant：Rust 实现、高性能、API 简洁、内置 payload 过滤、适合中等规模。Pinecone：托管、零运维、贵、不适合私有化。选型维度：数据规模、延迟要求、私有化需求、向量维度、混合检索（向量+过滤）。",
			tags: []string{"ai-agent", "vector-db", "milvus"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "什么是 Embedding 模型？如何选择合适的 embedding 模型？",
			a:    "Embedding 模型将文本映射为稠密向量，用于语义搜索/聚类/相似度计算。选型维度：1) 维度（维度高精度好但计算/存储贵）；2) 语言支持（多语言 vs 英文专用）；3) 最大 token 长度；4) MTEB benchmark 分数；5) 推理速度/成本。常用：text-embedding-v3（DashScope）、text-embedding-3-small（OpenAI）、BGE 系列（开源）。",
			tags: []string{"ai-agent", "embedding", "rag"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "什么是 Agentic Workflow？和单次 LLM 调用有什么区别？",
			a:    "Agentic Workflow 是多步骤、多工具调用的自动化流程，LLM 具有规划、执行、观察、修正的能力循环（Plan-Act-Observe）。区别：1) 多轮而非单次；2) 有外部工具调用；3) 有状态（记忆/上下文积累）；4) 可能有多个专门化子 Agent 协作（Multi-Agent）。",
			tags: []string{"ai-agent", "workflow", "multi-agent"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "LLM 幻觉（Hallucination）的原因和缓解手段？",
			a:    "原因：训练数据噪声、RLHF 过度自信、知识截止、提示词歧义。缓解：1) RAG 提供事实依据；2) 降低温度；3) 要求输出引用来源；4) 多模型投票（majority voting）；5) 验证层（另一个 LLM 或规则检查输出）；6) 微调（SFT/RLHF）减少错误模式。",
			tags: []string{"ai-agent", "hallucination", "llm"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "什么是 RLHF（人类反馈强化学习）？SFT/RM/PPO 各自的作用？",
			a:    "RLHF 三阶段：1) SFT（有监督微调）：在示范数据上微调 LLM，建立基本指令遵循能力；2) RM（奖励模型）：用人类偏好对比数据训练打分模型；3) PPO：用 RM 的奖励信号优化 LLM 策略（约束 KL 散度防止奖励黑客）。目标是让模型输出符合人类偏好的高质量回答。",
			tags: []string{"ai-agent", "rlhf", "llm", "training"},
			diff: domain.DifficultyHard,
		},
		{
			q:    "描述 LangChain 或 Eino 等框架的核心抽象（Chain/Graph/Node）。",
			a:    "Chain/Graph：有向图描述数据流，节点（Node/Component）是独立处理单元（LLM/工具/检索器），边定义执行顺序或条件路由。核心抽象：Runnable（统一接口）、Context 传递状态、流式输出（streaming）。图结构使得复杂 Agent 流程可组合、可测试、可观测。",
			tags: []string{"ai-agent", "langchain", "eino", "framework"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "如何评估 RAG 系统的质量？有哪些常用指标？",
			a:    "检索侧：Recall@K（召回率）、MRR（平均倒数排名）、NDCG（归一化折损累积增益）。生成侧：Faithfulness（答案基于上下文，非幻觉）、Answer Relevancy（答案相关性）、Context Precision/Recall（RAGAS 框架）。端到端：人工标注 QA 对比、A/B 测试用户满意度。",
			tags: []string{"ai-agent", "rag", "evaluation"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "什么是 Token 限制问题？如何处理超长上下文？",
			a:    "LLM 有上下文窗口限制（如 32K/128K tokens）。处理策略：1) 截断（保留最近 N 轮）；2) 摘要压缩（定期将历史摘要）；3) RAG（检索而非全量放入）；4) 分块处理（MapReduce 分批 → 合并）；5) 使用长上下文模型（Gemini 1M / Claude 200K）但成本高。",
			tags: []string{"ai-agent", "context-window", "llm"},
			diff: domain.DifficultyMedium,
		},
		{
			q:    "Multi-Agent 系统中如何处理 Agent 间的协调与通信？",
			a:    "常见模式：1) Orchestrator-Worker（Supervisor 分解任务分配给专门 Agent）；2) Peer-to-Peer（Agent 直接消息传递）；3) Shared Memory（黑板系统，共享状态）；4) Event-Driven（通过消息队列解耦）。协调要解决：任务分配、状态同步、冲突检测、结果聚合。",
			tags: []string{"ai-agent", "multi-agent", "coordination"},
			diff: domain.DifficultyHard,
		},
		{
			q:    "如何对 LLM 应用做可观测性（Observability）？",
			a:    "核心三要素：Traces（每次 LLM 调用链路，含 prompt/response/latency/tokens）、Metrics（成功率/延迟 P99/token 消耗/成本）、Logs（结构化日志含 request_id）。工具：LangSmith（LangChain）、Phoenix（Arize）、Helicone、自建 OpenTelemetry + Jaeger。关键：trace_id 贯穿整个 Agent 调用链。",
			tags: []string{"ai-agent", "observability", "llm", "tracing"},
			diff: domain.DifficultyMedium,
		},
	}

	for _, q := range aiAgent {
		questions = append(questions, &domain.BankQuestionRecord{
			Question:            q.q,
			StandardAnswer:      q.a,
			Tags:                q.tags,
			RelatedConcepts:     []string{},
			FollowupQuestionIDs: []string{},
			Difficulty:          q.diff,
		})
	}

	return questions
}
