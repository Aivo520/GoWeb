
//界面刷新时弹出加载该界面
window.onload = function() {
    var data = { site: "127.0.0.1"}
    var vm = new Vue( {
        el: '#site',
        data: data
    })
};