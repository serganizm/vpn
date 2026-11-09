package com.serganizm.vpnmanager

import android.annotation.SuppressLint
import android.app.Activity
import android.content.Intent
import android.net.Uri
import android.os.Bundle
import android.provider.OpenableColumns
import android.webkit.JavascriptInterface
import android.webkit.ValueCallback
import android.webkit.WebChromeClient
import android.webkit.WebResourceRequest
import android.webkit.WebResourceResponse
import android.webkit.WebSettings
import android.webkit.WebView
import android.webkit.WebViewClient
import androidx.activity.result.contract.ActivityResultContracts
import androidx.appcompat.app.AppCompatActivity
import androidx.core.view.WindowCompat
import androidx.webkit.WebViewAssetLoader
import org.json.JSONObject
import java.nio.charset.Charset

class MainActivity : AppCompatActivity() {
    private lateinit var webView: WebView
    private val api = LocalApi(BuildConfig.VERSION_NAME)
    private var filePathCallback: ValueCallback<Array<Uri>>? = null

    private val fileChooserLauncher =
        registerForActivityResult(ActivityResultContracts.StartActivityForResult()) { result ->
            val callback = filePathCallback
            filePathCallback = null
            if (callback == null) return@registerForActivityResult
            val uris = if (result.resultCode == Activity.RESULT_OK) {
                WebChromeClient.FileChooserParams.parseResult(result.resultCode, result.data)
            } else null
            callback.onReceiveValue(uris)
        }

    private val nativeOvpnPicker =
        registerForActivityResult(ActivityResultContracts.OpenDocument()) { uri ->
            if (uri == null) {
                deliverNativePick(null, null, "cancelled")
                return@registerForActivityResult
            }
            try {
                contentResolver.takePersistableUriPermission(
                    uri,
                    Intent.FLAG_GRANT_READ_URI_PERMISSION
                )
            } catch (_: SecurityException) {
                // Not always persistable; ignore.
            }
            val name = queryDisplayName(uri) ?: "profile.ovpn"
            val text = contentResolver.openInputStream(uri)?.use { input ->
                input.readBytes().toString(Charset.forName("UTF-8"))
            } ?: ""
            deliverNativePick(name, text, null)
        }

    @SuppressLint("SetJavaScriptEnabled")
    override fun onCreate(savedInstanceState: Bundle?) {
        super.onCreate(savedInstanceState)
        WindowCompat.setDecorFitsSystemWindows(window, true)

        webView = WebView(this).apply {
            setBackgroundColor(0xFF090911.toInt())
        }
        setContentView(webView)

        val assetLoader = WebViewAssetLoader.Builder()
            .addPathHandler("/assets/", WebViewAssetLoader.AssetsPathHandler(this))
            .build()

        webView.settings.apply {
            javaScriptEnabled = true
            domStorageEnabled = true
            databaseEnabled = true
            allowFileAccess = true
            allowContentAccess = true
            mediaPlaybackRequiresUserGesture = false
            cacheMode = WebSettings.LOAD_DEFAULT
            mixedContentMode = WebSettings.MIXED_CONTENT_COMPATIBILITY_MODE
            useWideViewPort = true
            loadWithOverviewMode = true
            textZoom = 100
        }

        webView.webViewClient = object : WebViewClient() {
            override fun shouldInterceptRequest(
                view: WebView?,
                request: WebResourceRequest
            ): WebResourceResponse? {
                val url = request.url ?: return null
                if (url.host == "appassets.androidplatform.net") {
                    return assetLoader.shouldInterceptRequest(url)
                }
                return super.shouldInterceptRequest(view, request)
            }

            override fun onPageFinished(view: WebView?, url: String?) {
                super.onPageFinished(view, url)
                view?.evaluateJavascript(
                    """
                    (function(){
                      document.documentElement.classList.add('android-app');
                      document.body.classList.add('android-app');
                      try { if (window.__vpnAndroidReady) window.__vpnAndroidReady(); } catch (e) {}
                    })();
                    """.trimIndent(),
                    null
                )
            }
        }

        webView.webChromeClient = object : WebChromeClient() {
            override fun onShowFileChooser(
                webView: WebView?,
                filePathCallback: ValueCallback<Array<Uri>>?,
                fileChooserParams: FileChooserParams?
            ): Boolean {
                this@MainActivity.filePathCallback?.onReceiveValue(null)
                this@MainActivity.filePathCallback = filePathCallback
                val intent = try {
                    fileChooserParams?.createIntent()
                        ?: Intent(Intent.ACTION_GET_CONTENT).apply {
                            addCategory(Intent.CATEGORY_OPENABLE)
                            type = "*/*"
                        }
                } catch (_: Exception) {
                    Intent(Intent.ACTION_GET_CONTENT).apply {
                        addCategory(Intent.CATEGORY_OPENABLE)
                        type = "*/*"
                    }
                }
                return try {
                    fileChooserLauncher.launch(Intent.createChooser(intent, "Выберите .ovpn файл"))
                    true
                } catch (_: Exception) {
                    this@MainActivity.filePathCallback = null
                    filePathCallback?.onReceiveValue(null)
                    false
                }
            }
        }

        webView.addJavascriptInterface(ApiBridge(api) { openNativeOvpnPicker() }, "VpnNative")
        webView.loadUrl("https://appassets.androidplatform.net/assets/index.html")
    }

    private fun openNativeOvpnPicker() {
        nativeOvpnPicker.launch(arrayOf("*/*"))
    }

    private fun queryDisplayName(uri: Uri): String? {
        contentResolver.query(uri, arrayOf(OpenableColumns.DISPLAY_NAME), null, null, null)?.use { c ->
            if (c.moveToFirst()) {
                val idx = c.getColumnIndex(OpenableColumns.DISPLAY_NAME)
                if (idx >= 0) return c.getString(idx)
            }
        }
        return uri.lastPathSegment
    }

    private fun deliverNativePick(name: String?, content: String?, error: String?) {
        val payload = JSONObject()
            .put("name", name ?: JSONObject.NULL)
            .put("content", content ?: JSONObject.NULL)
            .put("error", error ?: JSONObject.NULL)
            .toString()
        val js = "window.__vpnNativeFile && window.__vpnNativeFile(" +
            JSONObject.quote(payload) + ");"
        webView.post {
            webView.evaluateJavascript(js, null)
        }
    }

    @Deprecated("Deprecated in Java")
    override fun onBackPressed() {
        if (::webView.isInitialized && webView.canGoBack()) {
            webView.goBack()
        } else {
            @Suppress("DEPRECATION")
            super.onBackPressed()
        }
    }

    override fun onDestroy() {
        filePathCallback?.onReceiveValue(null)
        filePathCallback = null
        webView.destroy()
        super.onDestroy()
    }
}

class ApiBridge(
    private val api: LocalApi,
    private val onPickOvpn: () -> Unit
) {
    @JavascriptInterface
    fun request(method: String, path: String, body: String?): String {
        val res = api.handle(method, path, body)
        return JSONObject()
            .put("code", res.code)
            .put("contentType", res.contentType)
            .put("body", res.body)
            .toString()
    }

    @JavascriptInterface
    fun pickOvpnFile() {
        onPickOvpn()
    }
}
