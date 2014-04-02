---
layout: post
title: Complex vs. Complexity
description: "Death by a thousand cuts: From complexity to complex."
tags: [go]
image:
  feature: abstract-3.jpg
  credit: dargadgetz
  creditlink: http://www.dargadgetz.com/ios-7-abstract-wallpaper-pack-for-iphone-5-and-ipod-touch-retina/
comments: true
share: true
---

I guess that every software developer will agree that *complex* systems are hard to maintain. They are hard to extend, hard to debug and hard to deploy. And I am not talking about badly engineered systems, aka "spaghetti code", but well written software that is crammed with too many features. So we want to avoid that. But how?

Helpful could be to make a clear distinction between "*complex*" and "*complexity*". When an application has become *complex*, it is too late. Recovery is time consuming because too many moving parts have to be considered. The solution has to be on the path to *complex*. That is where *complexity* comes in:

Each feature or function, as small as it might be, adds *complexity*. This is not inherently bad, so when we say that something adds *complexity*, it means that we recognize that we get a little closer to "*complex*". Each time that we want to add a new feature, we have to ask ourselves if the gains outweigh the extra *complexity* that comes with it. Big features that improve the product a lot allow many new lines of code and new sophisticated program structures. Still, in some cases such a feature gets dropped because "it is too *complex*". Here, the decision process is quite obvious because it would add so much *complexity* that we arrive directly at "*complex*". This is relatively easy to see and communicate.

Much more dangerous is the death by a thousand cuts: Especially for small to tiny features, we may fall to the false assumption that adding them has no long-term cost. It seems as if we have to spend some developer's time on implementing the feature, but when it is done, then our product is better off than before. This is not generally true, because we just added *complexity*, even if it was just a tiny bit. Let's say that for something to be *complex*, it needs 100 complexity points. The feature just mentioned is so small, it just adds 1 complexity point. Still, we are now 1 point closer to a hard to maintain system and, assuming that we started at zero, we now can only spend 99 remaining points on other maybe more important features. Repeat that with a lot of small features and you will reach the 100 faster that you thought, without any big feature and without it being immediately obvious.

Of course there are countermeasures, like splitting the system into multiple parts with clear boundaries, etc. There is much literature describing those best practices. What I intended with this article is to show that *complexity* is not bad per se, but a lot of it leads to a *complex* system and thus it has to be considered at many occasions.