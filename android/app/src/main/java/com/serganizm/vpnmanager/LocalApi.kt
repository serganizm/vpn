package com.serganizm.vpnmanager

import org.json.JSONArray
import org.json.JSONObject
import java.util.UUID
import java.util.concurrent.ConcurrentHashMap
import java.util.concurrent.Executors
import java.util.concurrent.atomic.AtomicLong

/**
 * In-app dry-run control API mirroring the desktop Go service.
 * Server 2 is managed only via GRE through Server 1 after bootstrap.
 */
class LocalApi(private val version: String = "0.1.0") {
    private val pairs = ConcurrentHashMap<String, JSONObject>()
    private val users = ConcurrentHashMap<Long, JSONObject>()
    private val profiles = ConcurrentHashMap<String, JSONObject>()
    private val jobs = ConcurrentHashMap<String, JSONObject>()
    private val nextUserId = AtomicLong(1)
    private val executor = Executors.newCachedThreadPool()

    private val steps = listOf(
        "instSsh1", "instPkg1", "instOvpn1", "instSsh2ViaS1", "instPkg2",
        "instOvpn2", "instGre", "instFirewallS2", "instUser", "instDns", "instCheck"
    )

    data class Response(val code: Int, val body: String, val contentType: String = "application/json")

    fun handle(method: String, path: String, body: String?): Response {
        val clean = path.substringBefore('?')
        return when {
            method == "GET" && clean == "/api/v1/info" -> ok(
                JSONObject()
                    .put("name", "VPN Manager")
                    .put("version", version)
                    .put("mode", "dry-run")
                    .put("protocol", "OpenVPN")
                    .put("platform", "android")
                    .put(
                        "notes",
                        JSONArray()
                            .put("Clients connect to Server 1 only")
                            .put("Server 2 is managed via GRE through Server 1")
                            .put("Public IP of Server 2 is never used for admin after bootstrap")
                    )
            )
            method == "GET" && clean == "/api/v1/pairs" -> ok(JSONArray(pairs.values.map { JSONObject(it.toString()) }))
            method == "POST" && clean == "/api/v1/install" -> startInstall(body)
            method == "GET" && clean.startsWith("/api/v1/install/") -> getJob(clean.removePrefix("/api/v1/install/"))
            method == "GET" && clean == "/api/v1/users" -> {
                val pair = path.substringAfter("pair=", "").substringBefore('&').ifBlank { null }
                val arr = JSONArray()
                users.values.filter { pair == null || it.optString("pair") == pair }.forEach { arr.put(JSONObject(it.toString())) }
                ok(arr)
            }
            method == "POST" && clean == "/api/v1/users" -> addUser(body)
            method == "GET" && clean == "/api/v1/profiles" -> ok(JSONArray(profiles.values.map { JSONObject(it.toString()) }))
            method == "POST" && clean == "/api/v1/profiles" -> addProfile(body)
            method == "GET" && clean.matches(Regex("/api/v1/users/\\d+/config\\.ovpn")) -> {
                val id = clean.split('/')[4].toLong()
                userConfig(id)
            }
            method == "POST" && clean.matches(Regex("/api/v1/pairs/[^/]+/settings")) -> {
                val id = clean.split('/')[4]
                applySettings(id, body)
            }
            method == "POST" && clean.matches(Regex("/api/v1/pairs/[^/]+/session")) -> {
                val id = clean.split('/')[4]
                setSession(id, body)
            }
            method == "GET" && clean.matches(Regex("/api/v1/pairs/[^/]+")) -> {
                val id = clean.split('/')[4]
                pairs[id]?.let { ok(it) } ?: notFound()
            }
            else -> Response(404, """{"error":"not found"}""")
        }
    }

    private fun startInstall(body: String?): Response {
        val req = JSONObject(body ?: "{}")
        val name = req.optString("name").trim()
        val servers = req.optJSONArray("servers")
        if (name.isEmpty() || servers == null || servers.length() < 2) {
            return Response(400, """{"error":"need name and two servers"}""")
        }
        val id = UUID.randomUUID().toString()
        val job = JSONObject()
            .put("id", id)
            .put("pairName", name)
            .put("index", 0)
            .put("steps", JSONArray(steps))
            .put("log", JSONArray())
            .put("done", false)
            .put("ok", false)
            .put("dryRun", true)
            .put("createdAt", java.time.Instant.now().toString())
        jobs[id] = job
        executor.execute { runInstall(id, req) }
        return ok(job)
    }

    private fun runInstall(jobId: String, req: JSONObject) {
        val fail = req.optString("name").contains("fail", ignoreCase = true)
        val servers = req.getJSONArray("servers")
        val s1 = servers.getJSONObject(0)
        val s2 = servers.getJSONObject(1)
        for ((i, step) in steps.withIndex()) {
            Thread.sleep(350)
            val job = jobs[jobId] ?: return
            if (fail && i == 3) {
                appendLog(job, "fail $step: simulated failure")
                job.put("done", true).put("ok", false).put("error", step)
                return
            }
            appendLog(job, "ok $step")
            if (i >= steps.lastIndex) {
                job.put("done", true).put("ok", true).put("index", steps.size)
                appendLog(job, "install complete")
                val pairId = "p" + UUID.randomUUID().toString().take(8)
                val pair = JSONObject()
                    .put("id", pairId)
                    .put("name", req.optString("name").trim())
                    .put("session", true)
                    .put("gre", true)
                    .put("dns", req.optString("dns", "google"))
                    .put("dnsCustom", req.optJSONArray("dnsCustom") ?: JSONArray().put("").put(""))
                    .put("route", req.optString("route", "split"))
                    .put("lists", JSONObject().put("iplist", JSONArray()).put("rueu", JSONArray()).put("ru", JSONArray()))
                    .put(
                        "a",
                        JSONObject()
                            .put("ip", s1.optString("ip"))
                            .put("port", s1.optString("port", "22"))
                            .put("login", s1.optString("login"))
                            .put("online", true)
                            .put("cpu", 2)
                            .put("ram", 10)
                            .put("hdd", 12)
                            .put("uptime", "0m")
                            .put("managementIp", s1.optString("ip"))
                    )
                    .put(
                        "b",
                        JSONObject()
                            .put("ip", s2.optString("ip"))
                            .put("port", s2.optString("port", "22"))
                            .put("login", s2.optString("login"))
                            .put("online", true)
                            .put("cpu", 1)
                            .put("ram", 8)
                            .put("hdd", 10)
                            .put("uptime", "0m")
                            .put("managementIp", "10.10.10.2")
                    )
                    .put("greLocal", "10.10.10.1")
                    .put("greRemote", "10.10.10.2")
                pairs[pairId] = pair
                val uid = nextUserId.getAndIncrement()
                users[uid] = JSONObject()
                    .put("id", uid)
                    .put("pair", pairId)
                    .put("name", req.optString("userName", "user").ifBlank { "user" })
                    .put("comment", req.optString("userComment"))
                    .put("auth", req.optString("auth", "file"))
                    .put("limit", req.optInt("limit", 0))
                    .put("connections", JSONArray())
                val pid = "c" + UUID.randomUUID().toString().take(8)
                profiles[pid] = JSONObject().put("id", pid).put("name", pair.getString("name"))
                job.put("pairId", pairId)
                return
            }
            job.put("index", i + 1)
        }
    }

    private fun appendLog(job: JSONObject, msg: String) {
        val stamp = java.time.LocalTime.now().withNano(0).toString()
        job.getJSONArray("log").put("[$stamp] $msg")
    }

    private fun getJob(id: String): Response = jobs[id]?.let { ok(it) } ?: notFound()

    private fun addUser(body: String?): Response {
        val req = JSONObject(body ?: "{}")
        val pairId = req.optString("pair")
        val pair = pairs[pairId] ?: return Response(404, """{"error":"pair not found"}""")
        if (!pair.optBoolean("session") || !pair.optBoolean("gre")) {
            return Response(409, """{"error":"Server 1 session + GRE required"}""")
        }
        val uid = nextUserId.getAndIncrement()
        val user = JSONObject()
            .put("id", uid)
            .put("pair", pairId)
            .put("name", req.optString("name", "user"))
            .put("comment", req.optString("comment"))
            .put("auth", req.optString("auth", "file"))
            .put("limit", req.optInt("limit", 0))
            .put("connections", JSONArray())
        users[uid] = user
        return ok(user)
    }

    private fun addProfile(body: String?): Response {
        val req = JSONObject(body ?: "{}")
        val id = req.optString("id").ifBlank { "c" + UUID.randomUUID().toString().take(8) }
        val p = JSONObject().put("id", id).put("name", req.optString("name", "VPN"))
        profiles[id] = p
        return ok(p)
    }

    private fun applySettings(id: String, body: String?): Response {
        val pair = pairs[id] ?: return notFound()
        if (!pair.optBoolean("session")) return Response(409, """{"error":"connect to Server 1 session first"}""")
        if (!pair.optBoolean("gre")) return Response(409, """{"error":"GRE is down; cannot manage Server 2 via tunnel"}""")
        val req = JSONObject(body ?: "{}")
        if (req.has("dns")) pair.put("dns", req.getString("dns"))
        if (req.has("dnsCustom")) pair.put("dnsCustom", req.getJSONArray("dnsCustom"))
        if (req.has("route")) pair.put("route", req.getString("route"))
        if (req.has("lists")) pair.put("lists", req.getJSONObject("lists"))
        return ok(
            JSONObject()
                .put("pair", pair)
                .put("appliedVia", pair.optString("greRemote", "10.10.10.2"))
                .put("viaJump", true)
                .put("managementNote", "Server 2 via GRE through Server 1 (public IP blocked / forbidden)")
                .put("forbiddenDirect", true)
        )
    }

    private fun setSession(id: String, body: String?): Response {
        val pair = pairs[id] ?: return notFound()
        val on = JSONObject(body ?: "{}").optBoolean("on", !pair.optBoolean("session"))
        pair.put("session", on)
        return ok(pair)
    }

    private fun userConfig(id: Long): Response {
        val user = users[id] ?: return notFound()
        val pair = pairs[user.getString("pair")] ?: return notFound()
        val remote = pair.getJSONObject("a").getString("ip")
        val s2 = pair.getJSONObject("b").getString("ip")
        val cfg = buildString {
            appendLine("# VPN Manager ${user.getString("name")}")
            appendLine("# remote = Server 1 (entry). Server 2 is reachable only via GRE through S1.")
            appendLine("client")
            appendLine("dev tun")
            appendLine("proto udp")
            appendLine("remote $remote 1194")
            appendLine("resolv-retry infinite")
            appendLine("nobind")
            appendLine("persist-key")
            appendLine("persist-tun")
            appendLine("remote-cert-tls server")
            appendLine("cipher AES-256-GCM")
            appendLine("auth SHA256")
            appendLine("verb 3")
        }
        if (cfg.contains(s2)) {
            return Response(500, "refusing config that points at Server 2 public IP")
        }
        return Response(200, cfg, "application/x-openvpn-profile")
    }

    private fun ok(value: Any) = Response(200, value.toString())
    private fun notFound() = Response(404, """{"error":"not found"}""")
}
