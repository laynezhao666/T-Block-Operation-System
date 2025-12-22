/**
 * 星云业务播放器
 */
;
(function (undefined) {
    "use strict";
    var _global;

    /**
     * 工具函数 - 对象合并
     * @param o 旧对象
     * @param n 新对象
     * @param override 覆盖
     * @returns {*}
     */
    function extend(o, n, override) {
        for (var key in n) {
            if (n.hasOwnProperty(key) && (!o.hasOwnProperty(key) || override)) {
                o[key] = n[key];
            }
        }
        return o;
    }

    /**
     * 工具函数--转json  截取自json2
     * (utf8无bom格式)
     */
    function toJson(text) {
        var rx_one = /^[\],:{}\s]*$/;
        var rx_two = /\\(?:["\\\/bfnrt]|u[0-9a-fA-F]{4})/g;
        var rx_three = /"[^"\\\n\r]*"|true|false|null|-?\d+(?:\.\d*)?(?:[eE][+\-]?\d+)?/g;
        var rx_four = /(?:^|:|,)(?:\s*\[)+/g;
        var rx_escapable = /[\\"\u0000-\u001f\u007f-\u009f\u00ad\u0600-\u0604\u070f\u17b4\u17b5\u200c-\u200f\u2028-\u202f\u2060-\u206f\ufeff\ufff0-\uffff]/g;
        var rx_dangerous = /[\u0000\u00ad\u0600-\u0604\u070f\u17b4\u17b5\u200c-\u200f\u2028-\u202f\u2060-\u206f\ufeff\ufff0-\uffff]/g;
        text = String(text);
        rx_dangerous.lastIndex = 0;
        if (rx_dangerous.test(text)) {
            text = text.replace(rx_dangerous, function (a) {
                return (
                    "\\u"
                    + ("0000" + a.charCodeAt(0).toString(16)).slice(-4)
                );
            });
        }
        if (
            rx_one.test(
                text
                    .replace(rx_two, "@")
                    .replace(rx_three, "]")
                    .replace(rx_four, "")
            )
        ) {
            var j = eval("(" + text + ")");
            return j;
        }
        alert('数据格式错误');
        throw new SyntaxError("JSON error");
    }

    /**
     * 工具函数-ajax
     */
    function request(opt) {
        // 默认参数
        var def = {
            async: true,
            type: 'GET',
            url: '/video/',
            data: '',
            success: function () {
            },
            error: function () {
            }
        };
        def = extend(def, opt, true); //配置参数
        var xmlhttp;
        if (window.XMLHttpRequest) {
            //  IE7+, Firefox, Chrome, Opera, Safari 浏览器执行代码
            xmlhttp = new XMLHttpRequest();
        }
        else {
            // IE6, IE5 浏览器执行代码
            xmlhttp = new ActiveXObject("Microsoft.XMLHTTP");
        }
        xmlhttp.onreadystatechange = function () {
            if (xmlhttp.readyState == 4) {
                if (xmlhttp.status == 200) {
                    def.success(xmlhttp.responseText);
                } else {
                    def.error();
                    console.error(xmlhttp);
                }
            }
        }
        xmlhttp.open(def.type, def.url, def.async);
        xmlhttp.setRequestHeader("Content-type", "application/x-www-form-urlencoded");
        xmlhttp.send(def.data);
    }

    /**
     * 工具函数-
     * @param e //url
     * @param t //回调
     * @param i //属性
     */
    function loadScript(e, t, i) {
        var o = arguments.length > 3 && void 0 !== arguments[3] && arguments[3],
            n = document.createElement("script");
        if (n.onload = n.onreadystatechange = function () {
                this.readyState 
                && "loaded" !== this.readyState 
                && "complete" !== this.readyState 
                || ("function" == typeof t 
                    && t(), 
                    n.onload = n.onreadystatechange = null, 
                    n.parentNode && !o && n.parentNode.removeChild(n)
                    )
            },
                i) for (var r in i) if (i.hasOwnProperty(r)) {
            var s = i[r];
            null === s ? n.removeAttribute(s) : n.setAttribute(r, s)
        }
        n.src = e,
            document.getElementsByTagName("head")[0].appendChild(n)
    }

    /**
     * 通过class查找dom
     */
    if (!('getElementsByClass' in HTMLElement)) {
        HTMLElement.prototype.getElementsByClass = function (n) {
            var el = [],
                _el = this.getElementsByTagName('*');
            for (var i = 0; i < _el.length; i++) {
                if (!!_el[i].className && (typeof _el[i].className == 'string') && _el[i].className.indexOf(n) > -1) {
                    el[el.length] = _el[i];
                }
            }
            return el;
        };
        ((typeof HTMLDocument !== 'undefined') ? HTMLDocument : Document).prototype.getElementsByClass = HTMLElement.prototype.getElementsByClass;
    }


    // 插件构造函数 - 返回数组结构
    /**
     *
     * @param id
     * @param extendOpt
     * @param playerOpt
     * @constructor
     */
    function NebulaPlayer(videoId, playerOpt, extendOpt, callback) {
        this._init(videoId, playerOpt, extendOpt, callback)
    }

    NebulaPlayer.prototype = {
        constructor: this,
        _init: function (videoId, playerOpt, extendOpt, callback) {
            var tcplayerUrl = '//imgcache.qq.com/open/qcloud/video/vcplayer/TcPlayer-2.2.2.js';
            var staticUrl = '//res.tnebula.cn/static/nebulaplayer'; //'//res.tnebula.cn/static/nebulaplayer';
            this.instance = null;  //播放器实例
            this.videoId = videoId;
            this.videoDiv = document.querySelector("#" + videoId);
            //this.hasInit= false; //检查是否已初始化

            //事件监听-观察者模式
            this.listeners = []; //自定义事件，用于监听插件的用户交互
            this.handlers = {};  //回调函数的集合

            // tcplayer播放器的默认参数
            this.def = {
                "autoplay": true,      //iOS下safari浏览器，以及大部分移动端浏览器是不开放视频自动播放这个能力的
                "flash": false,
                "h5_flv": true,
                //"live": true, //直播
                "volume": '0',
                //"width": '480',//视频的显示宽度，请尽量使用视频分辨率宽度
                //"height": '320',//视频的显示高度，请尽量使用视频分辨率高度
                "wording": {
                    404: "视频不存在",
                    4: "视频不存在",
                    2: "视频不存在"
                }
            };

            // nebulaplayer组件-默认参数
            this.extendDef = {
                'hasClose': true,
                //'hasDownLoad': true,
                'hasCutImg': true,
                'hasVolume': false,
                'hasSpeedChangeControl': false,
                'tcplayerUrl': tcplayerUrl,
                'staticUrl': staticUrl
            };

            this.loadOverrideStyle()

            var _this = this
            if (window.TcPlayer) {
                _this.setVideo(playerOpt, extendOpt, callback);
                return
            }
            setTimeout(function(){
                _this.videoDiv = document.querySelector("#" + videoId);
                var isLock = _this.videoDiv.getAttribute('isLock');
                if(isLock == null || isLock == 0){
                    _this.videoDiv.setAttribute('isLock',1);
                    //引入Tcplayer
                    loadScript(_this.extendDef.tcplayerUrl, function () {
                        _this.setVideo(playerOpt, extendOpt, callback);
                        var removeLogInterval = setInterval(function () {
                            if (!window.flvjs) {
                                return
                            }
                            clearInterval(removeLogInterval)
                            flvjs.LoggingControl.enableAll = false
                        }, 100)
                    }, {id: 'TcPlayer'});
                }
            },0)

        },

        loadOverrideStyle () {
            let styleElt = document.querySelector('#tnebula-player-style')
            if (styleElt) return

            styleElt = document.createElement('style')
            styleElt.innerHTML = `
                .vcp-fullscreen-toggle {
                    background-repeat: no-repeat;
                    background-position: center;
                    background-size: 28px;
                    width: 48px;
                    height: 48px;
                    padding: 0;
                    background-image: url("data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiB2aWV3Qm94PSIwIDAgMzYgMzYiIHZlcnNpb249IjEuMSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxuczp4bGluaz0iaHR0cDovL3d3dy53My5vcmcvMTk5OS94bGluayI+DQogICAgPHBhdGggZD0iTTcsMTYgTDEwLDE2IEwxMCwxMyBMMTMsMTMgTDEzLDEwIEw3LDEwIEw3LDE2IFoiIG9wYWN0aXk9IjEiIGZpbGw9IiNmZmYiPjwvcGF0aD4NCiAgICA8cGF0aCBkPSJNMjMsMTAgTDIzLDEzIEwyNiwxMyBMMjYsMTYgTDI5LDE2IEwyOSwxMCBMMjMsMTAgWiIgb3BhY3RpeT0iMSIgZmlsbD0iI2ZmZiI+PC9wYXRoPg0KICAgIDxwYXRoIGQ9Ik0yMywyMyBMMjMsMjYgTDI5LDI2IEwyOSwyMCBMMjYsMjAgTDI2LDIzIEwyMywyMyBaIiBvcGFjdGl5PSIxIiBmaWxsPSIjZmZmIj48L3BhdGg+DQogICAgPHBhdGggZD0iTTEwLDIwIEw3LDIwIEw3LDI2IEwxMywyNiBMMTMsMjMgTDEwLDIzIEwxMCwyMCBaIiBvcGFjdGl5PSIxIiBmaWxsPSIjZmZmIj48L3BhdGg+DQo8L3N2Zz4=") !important;
                }
                .vcp-fullscreen-toggle.cut-img {
                    background-repeat: no-repeat;
                    background-image: url("data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBzdGFuZGFsb25lPSJubyI/PjwhRE9DVFlQRSBzdmcgUFVCTElDICItLy9XM0MvL0RURCBTVkcgMS4xLy9FTiIgImh0dHA6Ly93d3cudzMub3JnL0dyYXBoaWNzL1NWRy8xLjEvRFREL3N2ZzExLmR0ZCI+PHN2ZyB0PSIxNTUxNjk4NjUxNDgwIiBjbGFzcz0iaWNvbiIgc3R5bGU9IiIgdmlld0JveD0iMCAwIDEwMjQgMTAyNCIgdmVyc2lvbj0iMS4xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHAtaWQ9IjQxNjUiIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB3aWR0aD0iNDgiIGhlaWdodD0iNDgiPjxkZWZzPjxzdHlsZSB0eXBlPSJ0ZXh0L2NzcyI+PC9zdHlsZT48L2RlZnM+PHBhdGggZD0iTTg4Ny40NjY2NjcgMjY0LjUzMzMzM0gxMzYuNTMzMzMzYy0yOC4xNiAwLTUxLjIgMjMuMDQtNTEuMiA1MS4ydjQ5NC45MzMzMzRjMCAyOC4xNiAyMy4wNCA1MS4yIDUxLjIgNTEuMmg3NTAuOTMzMzM0YzI4LjE2IDAgNTEuMi0yMy4wNCA1MS4yLTUxLjJWMzE1LjczMzMzM2MwLTI4LjE2LTIzLjA0LTUxLjItNTEuMi01MS4yeiBtLTM3NS40NjY2NjcgNTEyYy0xMTcuNzYgMC0yMTMuMzMzMzMzLTk1LjU3MzMzMy0yMTMuMzMzMzMzLTIxMy4zMzMzMzNzOTUuNTczMzMzLTIxMy4zMzMzMzMgMjEzLjMzMzMzMy0yMTMuMzMzMzMzIDIxMy4zMzMzMzMgOTUuNTczMzMzIDIxMy4zMzMzMzMgMjEzLjMzMzMzMy05NS41NzMzMzMgMjEzLjMzMzMzMy0yMTMuMzMzMzMzIDIxMy4zMzMzMzN6IG0yOTAuMTMzMzMzLTMyNC4yNjY2NjZjLTI4LjE2IDAtNTEuMi0yMy4wNC01MS4yLTUxLjJzMjMuMDQtNTEuMiA1MS4yLTUxLjIgNTEuMiAyMy4wNCA1MS4yIDUxLjItMjMuMDQgNTEuMi01MS4yIDUxLjJ6IiBmaWxsPSIjZmZmZmZmIiBwLWlkPSI0MTY2Ij48L3BhdGg+PHBhdGggZD0iTTUxMiA0MTguMTMzMzMzYy03OS43ODY2NjcgMC0xNDUuMDY2NjY3IDY1LjI4LTE0NS4wNjY2NjcgMTQ1LjA2NjY2N3M2NS4yOCAxNDUuMDY2NjY3IDE0NS4wNjY2NjcgMTQ1LjA2NjY2NyAxNDUuMDY2NjY3LTY1LjI4IDE0NS4wNjY2NjctMTQ1LjA2NjY2Ny02NS4yOC0xNDUuMDY2NjY3LTE0NS4wNjY2NjctMTQ1LjA2NjY2N3pNNzI1LjMzMzMzMyAyNjQuNTMzMzMzVjIxMy4zMzMzMzNjMC0yOC4xNi0yMy4wNC01MS4yLTUxLjItNTEuMkgzNDkuODY2NjY3Yy0yOC4xNiAwLTUxLjIgMjMuMDQtNTEuMiA1MS4ydjUxLjJoNDI2LjY2NjY2NnoiIGZpbGw9IiNmZmZmZmYiIHAtaWQ9IjQxNjciPjwvcGF0aD48L3N2Zz4=") !important;
                }
                .vcp-fullscreen-toggle.download {
                    background-repeat: no-repeat;
                    background-image: url("data:image/svg+xml;base64,PD94bWwgdmVyc2lvbj0iMS4wIiBzdGFuZGFsb25lPSJubyI/PjwhRE9DVFlQRSBzdmcgUFVCTElDICItLy9XM0MvL0RURCBTVkcgMS4xLy9FTiIgImh0dHA6Ly93d3cudzMub3JnL0dyYXBoaWNzL1NWRy8xLjEvRFREL3N2ZzExLmR0ZCI+PHN2ZyB0PSIxNTUxNjk4NzU3NjIyIiBjbGFzcz0iaWNvbiIgc3R5bGU9IiIgdmlld0JveD0iMCAwIDEwMjQgMTAyNCIgdmVyc2lvbj0iMS4xIiB4bWxucz0iaHR0cDovL3d3dy53My5vcmcvMjAwMC9zdmciIHAtaWQ9IjUzMjkiIHhtbG5zOnhsaW5rPSJodHRwOi8vd3d3LnczLm9yZy8xOTk5L3hsaW5rIiB3aWR0aD0iNDgiIGhlaWdodD0iNDgiPjxkZWZzPjxzdHlsZSB0eXBlPSJ0ZXh0L2NzcyI+PC9zdHlsZT48L2RlZnM+PHBhdGggZD0iTTk2Ny4xMTExMTEgOTY3LjExMTExMUg1Ni44ODg4ODl2LTI4NC40NDQ0NDRoMTEzLjc3Nzc3OHYxNzAuNjY2NjY2aDY4Mi42NjY2NjZ2LTE3MC42NjY2NjZoMTEzLjc3Nzc3OHoiIGZpbGw9IiNmZmZmZmYiIHAtaWQ9IjUzMzAiPjwvcGF0aD48cGF0aCBkPSJNNTEyIDc3OS4zNzc3NzhsLTI3OC43NTU1NTYtMjg0LjQ0NDQ0NSA3OS42NDQ0NDUtNzkuNjQ0NDQ0TDUxMiA2MTQuNGwxOTkuMTExMTExLTE5OS4xMTExMTEgNzkuNjQ0NDQ1IDc5LjY0NDQ0NHoiIGZpbGw9IiNmZmZmZmYiIHAtaWQ9IjUzMzEiPjwvcGF0aD48cGF0aCBkPSJNNTY4Ljg4ODg4OSA2MjUuNzc3Nzc4SDQ1NS4xMTExMTFWNTYuODg4ODg5aDExMy43Nzc3Nzh6IiBmaWxsPSIjZmZmZmZmIiBwLWlkPSI1MzMyIj48L3BhdGg+PC9zdmc+") !important;
                }
                .vcp-playtoggle {
                    background-image: url("data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiB2aWV3Qm94PSIwIDAgMzYgMzYiIHZlcnNpb249IjEuMSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxuczp4bGluaz0iaHR0cDovL3d3dy53My5vcmcvMTk5OS94bGluayI+DQogICAgPHBhdGggZD0iTTExLDEwIEwxOCwxMy43NCAxOCwyMi4yOCAxMSwyNiBNMTgsMTMuNzQgTDI2LDE4IDI2LDE4IDE4LDIyLjI4IiBmaWxsPSIjZmZmIj48L3BhdGg+DQo8L3N2Zz4=") !important;
                }
                .vcp-playing .vcp-playtoggle {
                    background-image: url("data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiB2aWV3Qm94PSIwIDAgMzYgMzYiIHZlcnNpb249IjEuMSIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIiB4bWxuczp4bGluaz0iaHR0cDovL3d3dy53My5vcmcvMTk5OS94bGluayI+DQogICAgPHBhdGggZD0iTTExLDEwIEwxNywxMCAxNywyNiAxMSwyNiBNMjAsMTAgTDI2LDEwIDI2LDI2IDIwLDI2IiBmaWxsPSIjZmZmIj48L3BhdGg+DQo8L3N2Zz4=") !important;
                }
            `
            document.body.append(styleElt)
        },

        //获取视频相关信息，并创建实例
        setVideo: function (playerOpt, extendOpt, callback) {
            this.def = extend(this.def, playerOpt, true); //设置参数
            this.extendDef = extend(this.extendDef, extendOpt, true); //设置参数
            var _this = this
            if (extendOpt.mp4 && extendOpt.mp4 != '') {
                _this.def.mp4 = extendOpt.mp4
                _this.loadVideo(_this.def)
                callback && callback();
            } else if (playerOpt.flv && playerOpt.flv != '') {
                _this.def.flv = playerOpt.flv
                _this.loadVideo(_this.def)
                callback && callback();
            } else {
                //根据_this.extendDef参数，获取视频的key和url
                var videoInfo = this.getVideoInfo()
                if (!videoInfo) {
                    return false
                } else {
                    // ajax获取token
                    var postData = 'key=' + videoInfo.urlKey
                    var request_data = request({
                        type: 'POST',
                        data: postData,
                        url: _this.extendDef.getTokenUrl + '?key=' + videoInfo.urlKey,
                        success: function (ret) {
                            ret = toJson(ret)
                            _this.def[videoInfo.realVideoType] = videoInfo.videoUrl + '?' + (ret.data.token || ret.data)
                            if (videoInfo.realVideoType == 'mp4') {
                                delete  _this.def.h5_flv
                            }
                            _this.loadVideo(_this.def)
                            callback && callback();
                        },
                        error: function () {
                        }
                    })
                }
            }

        },

        //创建视频实例
        loadVideo: function (params) {
            const that = this
            if (this.instance != null) {
                this.close();
            }

            const listener = (function (listener, player) {
                return function (msg) {
                    if (listener) {
                        listener(msg)
                    }

                    if(msg.type === 'error') {
                        const defaultMaxRetryDelay = 180000 // 3min: 3 * 60 * 1000 = 180000
                        const retryDelay = Math.min((player.retryDelay || 1000) * 2, params.maxRetryDelay || defaultMaxRetryDelay)
                        player.retryDelay = retryDelay

                        if (msg.reason === 'MEDIA_ERR_SRC_NOT_SUPPORTED') {
                            player.retryDelay = 900000 // 15min: 15 * 60 * 1000
                        }
                        player.retryTimeoutId = window.setTimeout(function () {
                            player.instance && player.instance.load()
                        }, player.retryDelay)
                    }
                }
            })(params.listener, this)

            params.listener = listener

            this.instance = new TcPlayer(this.videoId, params);

            var videoElt = document.querySelectorAll("#" + this.videoId + ' video')[0]

            // 处理跨域
            if (!params.live) {
                // 这么处理的原因见链接：https://juejin.im/entry/5865f38e570c3500688944c2
                videoElt.src = videoElt.src
                videoElt.setAttribute("crossOrigin", '*');//解决跨域
            }

            var _this = this;
            var el = document.querySelector("#" + this.videoId + ' .vcp-player')
            if (_this.extendDef.hasClose) {
                //生成关闭div
                var closeDiv = document.createElement('div');
                closeDiv.setAttribute("class", "")
                closeDiv.innerHTML = '<div style="position: absolute;z-index: 1005;top:0;color:ghostwhite;cursor:pointer;right: 0"><img height="16x" src="' + this.extendDef.staticUrl + '/img/close3.svg"></div>';
                el.appendChild(closeDiv);

                //绑定关闭事件
                closeDiv.onclick = function () {
                    _this.close();
                    if (_this.handlers['close']) {
                        _this.emit('close', _this)
                        //如果参数里面有回调，执行参数的回调
                        !!_this.def.close && _this.def.close.call(this, _this.dom);
                    }
                }
            }
            if (!_this.extendDef.hasVolume) {
                document.querySelector("#" + this.videoId + ' .vcp-volume').style.display = 'none'
            }

            //截图功能
            if (_this.extendDef.hasCutImg) {
                var cutImgDiv = document.createElement('div');
                cutImgDiv.setAttribute("class", "vcp-fullscreen-toggle cut-img")
                cutImgDiv.setAttribute("style", 'background-image:url(' + this.extendDef.staticUrl + '/img/cut_img.svg);background-repeat:no-repeat;margin-right:0px;background-size:28px;float:right;')
                var parentNode = document.querySelector("#" + this.videoId + ' .vcp-controls-panel')
                parentNode.appendChild(cutImgDiv);
                cutImgDiv.onclick = function () {
                    var tag = _this.videoId;
                    var $video = document.querySelectorAll("#" + _this.videoId + ' video')[0];
                    // const $video = document.createElement('video')
                    var canvas = document.createElement("canvas");
                    var ctx = canvas.getContext('2d')
                    canvas.width = 1280;
                    canvas.height = canvas.width * 9 / 16;
                    $video.setAttribute("crossOrigin", '*');//解决跨域

                    ctx.drawImage($video, 0, 0, canvas.width, canvas.height);

                    var base64 = canvas.toDataURL('image/png');
                    /*耗时任务处理，使用web worker*/
                    _this.startWorker(function () {
                        window.cutvideo_worker && window.cutvideo_worker.postMessage({type: 'dataURItoBlob', params: {base64: base64, tag: tag}});
                    });
                }
            }

            //下载功能
            if (_this.extendDef.hasDownLoad) {
                var downloadUrl = '';
                if (params.flv && params.flv != '') {
                    downloadUrl = params.flv;
                }
                if (params.mp4 && params.mp4 != '') {
                    downloadUrl = params.mp4;
                }
                var downloadLinkElt = document.createElement('a');
                downloadLinkElt.setAttribute("class", "vcp-fullscreen-toggle download");
                downloadLinkElt.setAttribute("href", downloadUrl + '&isDownload=1');
                downloadLinkElt.setAttribute("download", "")
                downloadLinkElt.setAttribute('target', '_blank')
                downloadLinkElt.setAttribute("style", 'background-image:url(' + this.extendDef.staticUrl + '/img/download.svg);background-size:24px;background-repeat:no-repeat;float:right')
                var parentNode = document.querySelector("#" + this.videoId + ' .vcp-controls-panel')
                parentNode.appendChild(downloadLinkElt);
            }

            //播放速度功能
            if (_this.extendDef.hasSpeedChangeControl) {
                var speedDiv = document.createElement('div');
                speedDiv.setAttribute("style", 'position:relative;z-index:1001; background-color:#fff;margin-right:16px;background-size:28px;float:right;margin-top:14px;height:20px;line-height:20px;width:36px;text-align:center;border-radius: 2px;cursor:pointer;')
                var parentNode = document.querySelector("#" + this.videoId + ' .vcp-controls-panel')
                parentNode.appendChild(speedDiv);

                var currentSpeedSpan = document.createElement('span')
                currentSpeedSpan.innerText = 'X1';
                speedDiv.appendChild(currentSpeedSpan)

                var moveoutTimeoout = null
                var onmouseenter = function () {
                    clearTimeout(moveoutTimeoout)
                    speedOptsUl.style.display = 'block'
                }

                var onSppedClick = function (evt) {
                    var speed = evt.target.dataset.speed
                    if (!speed) return

                    currentSpeedSpan.innerText = 'X' + speed;

                    var $video = document.querySelectorAll("#" + _this.videoId + ' video')[0];

                    $video.playbackRate = speed
                }

                var speedOptsUl = document.createElement('ul')
                speedOptsUl.style = 'display: block;position: absolute;left: -8px;bottom: 24px;background: #fff;background: rgba(255, 255, 255, 0.8);width: 48px;'
                var speedList = _this.extendDef.speedList || [1, 1.5, 2]
                speedList.forEach(speed => {
                    var speedLi = document.createElement('li')
                    speedLi.dataset.speed = speed
                    speedLi.innerText = 'X' + speed
                    speedLi.style = 'border-bottom: 1px solid #000;cursor: pointer;'
                    speedOptsUl.appendChild(speedLi)
                    speedLi.onmouseenter = onmouseenter
                    speedLi.onclick = onSppedClick
                })
                speedDiv.appendChild(speedOptsUl)
                speedOptsUl.style.display = 'none'

                var moveoutTimeoout = null
                speedOptsUl.onmouseenter = currentSpeedSpan.onmouseenter = speedDiv.onmouseenter = onmouseenter
                speedDiv.onmouseout = function () {
                    moveoutTimeoout = setTimeout(function () {
                        speedOptsUl.style.display = 'none'
                    }, 300)
                }
            }
        },


        //根据_this.extendDef参数，获取视频的key和url
        getVideoInfo: function () {
            var _this = this
            var urlKey = ''
            var videoUrl = ''
            var realVideoType = 'flv';
            var urlType = '';
            if (!_this.extendDef.getTokenUrl) {
                console.error('getTokenUrl不能为空')
                return false
            }
            if (!_this.extendDef.cameraId) {
                console.error('cameraId不能为空')
                return false
            }
            if (!_this.extendDef.videoType) {
                console.error('videoType不能为空')
                return false
            }
            //实时类型
            if (_this.extendDef.videoType < 10) {
                //常规
                if (_this.extendDef.videoType == 1) {
                    urlKey = 'internal/' + _this.extendDef.cameraId
                }
                //跟踪
                else if (_this.extendDef.videoType == 2) {
                    urlKey = 'trace/' + _this.extendDef.cameraId
                } else {
                    console.error('未知videoType参数')
                    return false
                }
                //默认进度条
                _this.def.live = (typeof _this.def.live !== 'undefined') ? _this.def.live : true;
                //默认不要下载
                _this.extendDef.hasDownLoad = (typeof _this.extendDef.hasDownLoad !== 'undefined') ? _this.extendDef.hasDownLoad : false;
                videoUrl = _this.extendDef.videoBaseUrl + '/live/hfs/' + urlKey + '.flv'
            }
            //历史类型
            else {
                if (!_this.extendDef.time) {
                    console.error('当视频类型为历史（11、12、13、14等）的时候，time参数不可缺失')
                    return false
                }
                //常规历史
                var timestamp = parseInt((new Date()).valueOf() / 1000);
                if (_this.extendDef.videoType == 11 || _this.extendDef.videoType == 14) {
                    urlKey = 'internal/' + _this.extendDef.cameraId
                    //小于15分钟之前,使用flv，//大于15分钟，使用mp4
                    if (_this.extendDef.focusVideoType) {
                        realVideoType = _this.extendDef.focusVideoType
                        urlType = _this.extendDef.focusVideoType;
                    } else if (timestamp - _this.extendDef.time < 15 * 60) {
                        realVideoType = 'flv';
                        urlType = 'flv';
                    } else {
                        realVideoType = 'mp4'; //'mp4-syno'
                        urlType = (_this.extendDef.videoType == 11) ? 'mp4' : 'mp4-syno';
                    }
                }
                //跟踪历史
                else {
                    //跟踪历史（浓缩）
                    if (_this.extendDef.videoType == 12) {
                        urlKey = 'trace/' + _this.extendDef.cameraId
                        realVideoType = 'flv';
                        urlType = 'flv';
                    }
                    //跟踪历史（单人）
                    else if (_this.extendDef.videoType == 13) {
                        urlKey = 'trace/' + _this.extendDef.cameraId

                    } else if (_this.extendDef.videoType == 14) {
                        //urlType = 'mp4';
                        //realVideoType = 'mp4-syno';
                    } else {
                        console.error('未知videoType参数')
                        return false
                    }
                }
                videoUrl = _this.extendDef.videoBaseUrl + '/' + urlType + '/' + urlKey + '/' + _this.extendDef.time + '.' + realVideoType
                //设置非直播，去掉进度条
                _this.def.live = (typeof _this.def.live !== 'undefined') ? _this.def.live : false;
                //默认要下载
                _this.extendDef.hasDownLoad = (typeof _this.extendDef.hasDownLoad !== 'undefined') ? _this.extendDef.hasDownLoad : true;

            }
            return {
                urlKey: urlKey,
                videoUrl: videoUrl,
                realVideoType: realVideoType
            }
        },


        /**
         * 关闭并释放资源
         */
        close: function (callback) {
            if (this.instance != null) {
                //销毁flv实例
                if (this.instance.video && this.instance.video.flv) {
                    this.instance.video.flv.pause();
                    this.instance.video.flv.unload();
                    this.instance.video.flv.detachMediaElement();
                    this.instance.video.flv.destroy();
                }
                //销毁播放器实例
                this.instance.pause();
                this.instance.destroy();
                this.instance = null;
                if (this.retryTimeoutId !== undefined && this.retryTimeoutId !== null) {
                    clearTimeout(this.retryTimeoutId)
                }
                this.videoDiv && this.videoDiv.setAttribute('isLock',0)
            }
            //执行回调
            callback && callback();
        },


        //web worker多线程
        startWorker: function (callback) {
            if (typeof(Worker) !== "undefined") {
                if (typeof(window.cutvideo_worker) == "undefined") {
                    var _this = this;
                    //异步加载，实现跨域加载webworker
                    request({
                        type: 'GET',
                        data: {},
                        url: _this.extendDef.staticUrl + "/js/cutvideo.js",
                        success: function (ret) {
                            var blob = new Blob([ret]);
                            // Obtain a blob URL reference to our worker 'file'.
                            var blobURL = window.URL.createObjectURL(blob);
                            window.cutvideo_worker = new Worker(blobURL);
                            cutvideo_worker.onmessage = function (event) {
                                if (event.data.type && event.data.type == 'dataURItoBlob') {
                                    //接收处理结果
                                    var objectURL = window.URL.createObjectURL(event.data.params.blob);
                                    var link = document.createElement('a');
                                    link.style.display = 'none';
                                    link.href = objectURL;
                                    link.setAttribute('download', event.data.params.tag + '-' + Date.parse(new Date()) + '.jpg');
                                    document.body.appendChild(link);
                                    link.click();
                                    document.body.removeChild(link);
                                    window.URL.revokeObjectURL(objectURL);
                                } else {
                                    console.log('主线程收到参数');
                                    console.log(event.data);
                                }
                            };
                            callback && callback();
                        },
                        error: function () {
                        }
                    })
                }
                callback && callback();

            } else {
                console.log("抱歉，你的浏览器不支持 Web Workers...");
            }
        },
        stopWorker: function () {
            cutvideo_worker.terminate();
            cutvideo_worker = undefined;
        },


        // 自定义事件-监听（绑定）
        on: function (type, handler) {
            if (!this.handlers[type]) {
                this.handlers[type] = []
            }
            this.handlers[type].push(handler);
            this.listeners.push(type);
            return this;
        },

        // 解绑
        off: function (type, handler) {
            if (!this.handlers[type]) {
                return this;
            }
            if (this.handlers[type] && this.handlers[type] instanceof Array) {
                var len = this.handlers[type].length;
                for (var i = 0; i < len; i++) {
                    if (this.handlers[type][i] == handler) {
                        this.handlers[type].splice(i, 1);
                        this.listeners.splice(i, 1);
                    }
                }
            }
            return this;
        },

        once: function (type, handler) {

        },

        /**
         * 自定义事件的触发，实际上还是基于click之类的原生事件，我们只是把这种自定事件放到里面
         */
        emit: function () {
            //获取type,删除第一元素并返回
            var type = Array.prototype.shift.call(arguments)
            if (!this.handlers[type]) {
                return false;
            }
            if (this.handlers[type] && this.handlers[type] instanceof Array) {
                var len = this.handlers[type].length;
                for (var i = 0; i < len; i++) {
                    //执行函数
                    this.handlers[type][i].apply(this, arguments);
                }
            }
        },

    }

    // 最后将插件对象暴露给全局对象
    _global = (function () {
        return this || (0, eval)('this');
    }());
    if (typeof module !== "undefined" && module.exports) {
        module.exports = NebulaPlayer;
    } else if (typeof define === "function" && define.amd) {
        define(function () {
            return NebulaPlayer;
        });
    } else {
        !('NebulaPlayer' in _global) && (_global.NebulaPlayer = NebulaPlayer);
    }
}());
