const authInfo = {
    github: {
        scope: "user:email+read:user",
        endpoint: "https://github.com/login/oauth/authorize",
        clientId: "85db232fde2c9320ece7", // set your github client id
    },
    wechat: {
        scope: "snsapi_login",
        endpoint: "https://open.weixin.qq.com/connect/qrconnect",
        mpScope: "snsapi_userinfo",
        mpEndpoint: "https://open.weixin.qq.com/connect/oauth2/authorize",
        clientId: "",
        clientId2: ""
    },
};

const githubInfo = {
    project: 'https://github.com/eastygh/webm-nas',
    doc: 'https://github.com/eastygh/webm-nas/blob/master/README.md#weave',
}

export { authInfo, githubInfo };
