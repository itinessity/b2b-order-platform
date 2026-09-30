package com.eugenia.orders.domain;
public enum Market { MX("MXN"),CO("COP"),PE("PEN"); private final String currency; Market(String value){currency=value;} public String currency(){return currency;} }
