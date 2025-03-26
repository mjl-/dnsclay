package main

// WARNING: Automatically generated, do not edit manually. Add new providers to providers.txt, keeping it sorted, and run "make build".

import (
	"github.com/libdns/alidns"
	"github.com/libdns/autodns"
	"github.com/libdns/azure"
	"github.com/libdns/bunny"
	"github.com/libdns/civo"
	"github.com/libdns/cloudflare"
	"github.com/libdns/cloudns"
	"github.com/libdns/ddnss"
	"github.com/libdns/desec"
	"github.com/libdns/digitalocean"
	"github.com/libdns/directadmin"
	"github.com/libdns/dnsimple"
	"github.com/libdns/dnsmadeeasy"
	"github.com/libdns/dnspod"
	"github.com/libdns/dnsupdate"
	"github.com/libdns/domainnameshop"
	"github.com/libdns/dreamhost"
	"github.com/libdns/duckdns"
	"github.com/libdns/dynu"
	"github.com/libdns/dynv6"
	"github.com/libdns/easydns"
	"github.com/libdns/exoscale"
	"github.com/libdns/gandi"
	"github.com/libdns/gcore"
	"github.com/libdns/glesys"
	"github.com/libdns/godaddy"
	"github.com/libdns/googleclouddns"
	"github.com/libdns/he"
	"github.com/libdns/hetzner"
	"github.com/libdns/hexonet"
	"github.com/libdns/hosttech"
	"github.com/libdns/huaweicloud"
	"github.com/libdns/infomaniak"
	"github.com/libdns/inwx"
	"github.com/libdns/ionos"
	"github.com/libdns/katapult"
	"github.com/libdns/leaseweb"
	"github.com/libdns/linode"
	"github.com/libdns/loopia"
	"github.com/libdns/luadns"
	"github.com/libdns/mailinabox"
	"github.com/libdns/metaname"
	"github.com/libdns/mythicbeasts"
	"github.com/libdns/namecheap"
	"github.com/libdns/namedotcom"
	"github.com/libdns/namesilo"
	"github.com/libdns/nanelo"
	"github.com/libdns/netcup"
	"github.com/libdns/netlify"
	"github.com/libdns/nfsn"
	"github.com/libdns/njalla"
	"github.com/libdns/ovh"
	"github.com/libdns/porkbun"
	"github.com/libdns/powerdns"
	"github.com/libdns/rfc2136"
	"github.com/libdns/route53"
	"github.com/libdns/scaleway"
	"github.com/libdns/selectel"
	"github.com/libdns/tencentcloud"
	"github.com/libdns/timeweb"
	"github.com/libdns/totaluptime"
	"github.com/libdns/vultr"
)

// KnownProviders ensures all providers types are included in sherpadoc API documentation.
type KnownProviders struct {
	Xalidns         alidns.Provider
	Xautodns        autodns.Provider
	Xazure          azure.Provider
	Xbunny          bunny.Provider
	Xcivo           civo.Provider
	Xcloudflare     cloudflare.Provider
	Xcloudns        cloudns.Provider
	Xddnss          ddnss.Provider
	Xdesec          desec.Provider
	Xdigitalocean   digitalocean.Provider
	Xdirectadmin    directadmin.Provider
	Xdnsimple       dnsimple.Provider
	Xdnsmadeeasy    dnsmadeeasy.Provider
	Xdnspod         dnspod.Provider
	Xdnsupdate      dnsupdate.Provider
	Xdomainnameshop domainnameshop.Provider
	Xdreamhost      dreamhost.Provider
	Xduckdns        duckdns.Provider
	Xdynu           dynu.Provider
	Xdynv6          dynv6.Provider
	Xeasydns        easydns.Provider
	Xexoscale       exoscale.Provider
	Xgandi          gandi.Provider
	Xgcore          gcore.Provider
	Xglesys         glesys.Provider
	Xgodaddy        godaddy.Provider
	Xgoogleclouddns googleclouddns.Provider
	Xhe             he.Provider
	Xhetzner        hetzner.Provider
	Xhexonet        hexonet.Provider
	Xhosttech       hosttech.Provider
	Xhuaweicloud    huaweicloud.Provider
	Xinfomaniak     infomaniak.Provider
	Xinwx           inwx.Provider
	Xionos          ionos.Provider
	Xkatapult       katapult.Provider
	Xleaseweb       leaseweb.Provider
	Xlinode         linode.Provider
	Xloopia         loopia.Provider
	Xluadns         luadns.Provider
	Xmailinabox     mailinabox.Provider
	Xmetaname       metaname.Provider
	Xmythicbeasts   mythicbeasts.Provider
	Xnamecheap      namecheap.Provider
	Xnamedotcom     namedotcom.Provider
	Xnamesilo       namesilo.Provider
	Xnanelo         nanelo.Provider
	Xnetcup         netcup.Provider
	Xnetlify        netlify.Provider
	Xnfsn           nfsn.Provider
	Xnjalla         njalla.Provider
	Xovh            ovh.Provider
	Xporkbun        porkbun.Provider
	Xpowerdns       powerdns.Provider
	Xrfc2136        rfc2136.Provider
	Xroute53        route53.Provider
	Xscaleway       scaleway.Provider
	Xselectel       selectel.Provider
	Xtencentcloud   tencentcloud.Provider
	Xtimeweb        timeweb.Provider
	Xtotaluptime    totaluptime.Provider
	Xvultr          vultr.Provider
}

// providers is used for instantiating a provider by name.
var providers = map[string]any{
	"alidns":         alidns.Provider{},
	"autodns":        autodns.Provider{},
	"azure":          azure.Provider{},
	"bunny":          bunny.Provider{},
	"civo":           civo.Provider{},
	"cloudflare":     cloudflare.Provider{},
	"cloudns":        cloudns.Provider{},
	"ddnss":          ddnss.Provider{},
	"desec":          desec.Provider{},
	"digitalocean":   digitalocean.Provider{},
	"directadmin":    directadmin.Provider{},
	"dnsimple":       dnsimple.Provider{},
	"dnsmadeeasy":    dnsmadeeasy.Provider{},
	"dnspod":         dnspod.Provider{},
	"dnsupdate":      dnsupdate.Provider{},
	"domainnameshop": domainnameshop.Provider{},
	"dreamhost":      dreamhost.Provider{},
	"duckdns":        duckdns.Provider{},
	"dynu":           dynu.Provider{},
	"dynv6":          dynv6.Provider{},
	"easydns":        easydns.Provider{},
	"exoscale":       exoscale.Provider{},
	"gandi":          gandi.Provider{},
	"gcore":          gcore.Provider{},
	"glesys":         glesys.Provider{},
	"godaddy":        godaddy.Provider{},
	"googleclouddns": googleclouddns.Provider{},
	"he":             he.Provider{},
	"hetzner":        hetzner.Provider{},
	"hexonet":        hexonet.Provider{},
	"hosttech":       hosttech.Provider{},
	"huaweicloud":    huaweicloud.Provider{},
	"infomaniak":     infomaniak.Provider{},
	"inwx":           inwx.Provider{},
	"ionos":          ionos.Provider{},
	"katapult":       katapult.Provider{},
	"leaseweb":       leaseweb.Provider{},
	"linode":         linode.Provider{},
	"loopia":         loopia.Provider{},
	"luadns":         luadns.Provider{},
	"mailinabox":     mailinabox.Provider{},
	"metaname":       metaname.Provider{},
	"mythicbeasts":   mythicbeasts.Provider{},
	"namecheap":      namecheap.Provider{},
	"namedotcom":     namedotcom.Provider{},
	"namesilo":       namesilo.Provider{},
	"nanelo":         nanelo.Provider{},
	"netcup":         netcup.Provider{},
	"netlify":        netlify.Provider{},
	"nfsn":           nfsn.Provider{},
	"njalla":         njalla.Provider{},
	"ovh":            ovh.Provider{},
	"porkbun":        porkbun.Provider{},
	"powerdns":       powerdns.Provider{},
	"rfc2136":        rfc2136.Provider{},
	"route53":        route53.Provider{},
	"scaleway":       scaleway.Provider{},
	"selectel":       selectel.Provider{},
	"tencentcloud":   tencentcloud.Provider{},
	"timeweb":        timeweb.Provider{},
	"totaluptime":    totaluptime.Provider{},
	"vultr":          vultr.Provider{},
}
